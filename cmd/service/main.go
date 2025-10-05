package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/soheilhy/cmux"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	kafkalib "github.com/s21platform/kafka-lib"
	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	api "github.com/s21platform/materials-service/internal/generated"
	"github.com/s21platform/materials-service/internal/infra"
	"github.com/s21platform/materials-service/internal/pkg/tx"
	"github.com/s21platform/materials-service/internal/repository/postgres"
	"github.com/s21platform/materials-service/internal/rest"
	"github.com/s21platform/materials-service/internal/service"
	"github.com/s21platform/materials-service/pkg/materials"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()

	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Service.Name, cfg.Platform.Env)

	dbRepo := postgres.New(cfg)
	defer dbRepo.Close()

	editProducerConfig := kafkalib.DefaultProducerConfig(cfg.Kafka.Host, cfg.Kafka.Port, cfg.Kafka.EditMaterialTopic)

	editKafkaProducer := kafkalib.NewProducer(editProducerConfig)

	materialsService := service.New(dbRepo, editKafkaProducer)

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			infra.AuthInterceptorGRPC,
			infra.LoggerGRPC(logger),
			tx.TxMiddleWareGRPC(dbRepo),
		),
	)
	materials.RegisterMaterialsServiceServer(grpcServer, materialsService)

	createProducerConfig := kafkalib.DefaultProducerConfig(cfg.Kafka.Host, cfg.Kafka.Port, cfg.Kafka.MaterialCreatedTopic)

	createKafkaProducer := kafkalib.NewProducer(createProducerConfig)

	likeProducerConfig := kafkalib.DefaultProducerConfig(cfg.Kafka.Host, cfg.Kafka.Port, cfg.Kafka.LikeMaterialTopic)

	likeKafkaProducer := kafkalib.NewProducer(likeProducerConfig)

	handler := rest.New(dbRepo, createKafkaProducer, likeKafkaProducer)
	router := chi.NewRouter()

	router.Use(func(next http.Handler) http.Handler {
		return infra.AuthInterceptorHTTP(next)
	})
	router.Use(func(next http.Handler) http.Handler {
		return infra.LoggerHTTP(next, logger)
	})
	router.Use(func(next http.Handler) http.Handler {
		return tx.TxMiddlewareHTTP(dbRepo)(next)
	})

	api.HandlerFromMux(handler, router)
	httpServer := &http.Server{
		Handler: router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Service.Port))
	if err != nil {
		logger_lib.Error(context.Background(), fmt.Sprintf("failed to start TCP listener: %v", err))
	}

	m := cmux.New(listener)

	grpcListener := m.MatchWithWriters(cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	httpListener := m.Match(cmux.HTTP1Fast())

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		if err := grpcServer.Serve(grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			logger_lib.Error(logger_lib.WithError(ctx, err), "gRPC server error")
			return fmt.Errorf("gRPC server error: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := httpServer.Serve(httpListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger_lib.Error(logger_lib.WithError(ctx, err), "HTTP server error")
			return fmt.Errorf("HTTP server error: %v", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := m.Serve(); err != nil {
			logger_lib.Error(logger_lib.WithError(ctx, err), "cannot start service")
			return fmt.Errorf("cannot start service: %v", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "server error")
	}
}
