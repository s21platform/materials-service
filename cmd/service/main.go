package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	logger_lib "github.com/s21platform/logger-lib"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/infra"
	"github.com/s21platform/materials-service/internal/repository/postgres"
	"github.com/s21platform/materials-service/internal/service"
	"github.com/s21platform/materials-service/pkg/materials"
)

func main() {
	cfg := config.MustLoad()
	logger := logger_lib.New(cfg.Logger.Host, cfg.Logger.Port, cfg.Service.Name, cfg.Platform.Env)

	dbRepo := postgres.New(cfg)
	defer dbRepo.Close()

	materialsService := service.New(dbRepo)
	Server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			infra.AuthInterceptor,
			infra.Logger(logger),
		),
	)

	materials.RegisterMaterialsServiceServer(Server, materialsService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.Service.Port))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to start TCP listener: %v", err))
		log.Fatal("failed to start TCP listener: ", err)
	}

	if err = Server.Serve(listener); err != nil {
		logger.Error(fmt.Sprintf("failed to start gRPC listener: %v", err))
		log.Fatal("failed to start gRPC listener: ", err)
	}
}
