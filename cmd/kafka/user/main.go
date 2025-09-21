package main

import (
	"context"

	kafkalib "github.com/s21platform/kafka-lib"
	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/metrics-lib/pkg"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/databus/user"
	"github.com/s21platform/materials-service/internal/repository/postgres"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()

	dbRepo := postgres.New(cfg)
	defer dbRepo.Close()

	metrics, err := pkg.NewMetrics(cfg.Metrics.Host, cfg.Metrics.Port, cfg.Service.Name, cfg.Platform.Env)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to connect graphite:")
	}

	consumerConfig := kafkalib.DefaultConsumerConfig(
		cfg.Kafka.Host,
		cfg.Kafka.Port,
		cfg.Kafka.UserNicknameTopic,
		cfg.Kafka.UserNicknameConsumerGroup,
	)
	consumer, err := kafkalib.NewConsumer(consumerConfig, metrics)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to create consumer:")
	}

	userHandler := user.New(dbRepo)
	consumer.RegisterHandler(ctx, userHandler.Handler)

	<-ctx.Done()
}
