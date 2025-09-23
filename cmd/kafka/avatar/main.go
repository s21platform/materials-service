package main

import (
	"context"
	
	kafka_lib "github.com/s21platform/kafka-lib"
	logger_lib "github.com/s21platform/logger-lib"
	"github.com/s21platform/metrics-lib/pkg"

	"github.com/s21platform/materials-service/internal/config"
	"github.com/s21platform/materials-service/internal/databus/avatar"
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

	consumerConfig := kafka_lib.DefaultConsumerConfig(
		cfg.Kafka.Host,
		cfg.Kafka.Port,
		cfg.Kafka.AvatarTopic,
		cfg.Kafka.MaterialsAvatarUpdateKafkaConsumerGroup,
	)
	consumer, err := kafka_lib.NewConsumer(consumerConfig, metrics)
	if err != nil {
		logger_lib.Error(logger_lib.WithError(ctx, err), "failed to create consumer")
	}

	avatarHandler := avatar.New(dbRepo)
	consumer.RegisterHandler(ctx, avatarHandler.Handler)

	<-ctx.Done()
}
