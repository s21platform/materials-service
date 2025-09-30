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

	nicknameConsumerConfig := kafkalib.DefaultConsumerConfig(
		cfg.Kafka.Host,
		cfg.Kafka.Port,
		cfg.Kafka.UserNicknameTopic,
		cfg.Kafka.UserNicknameConsumerGroup,
	)

	userConsumerConfig := kafkalib.DefaultConsumerConfig(
		cfg.Kafka.Host,
		cfg.Kafka.Port,
		cfg.Kafka.UserTopic,
		cfg.Kafka.UserCreatedConsumerGroup,
	)

	nicknameConsumer, err := kafkalib.NewConsumer(nicknameConsumerConfig, metrics)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to create consumer:")
	}

	userConsumer, err := kafkalib.NewConsumer(userConsumerConfig, metrics)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to create consumer:")
	}

	userHandler := user.New(dbRepo)
	nicknameConsumer.RegisterHandler(ctx, userHandler.UpdateNickname)
	userConsumer.RegisterHandler(ctx, userHandler.UserCreated)

	<-ctx.Done()
}
