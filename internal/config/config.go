package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Service  Service
	Metrics  Metrics
	Platform Platform
	Postgres Postgres
	Logger   Logger
	Kafka    Kafka
}

type Service struct {
	Port string `env:"MATERIALS_SERVICE_PORT"`
	Name string `env:"MATERIALS_SERVICE_NAME"`
}

type Metrics struct {
	Host string `env:"GRAFANA_HOST"`
	Port int    `env:"GRAFANA_PORT"`
}

type Platform struct {
	Env string `env:"ENV"`
}

type Postgres struct {
	User     string `env:"MATERIALS_SERVICE_POSTGRES_USER"`
	Password string `env:"MATERIALS_SERVICE_POSTGRES_PASSWORD"`
	Database string `env:"MATERIALS_SERVICE_POSTGRES_DB"`
	Host     string `env:"MATERIALS_SERVICE_POSTGRES_HOST"`
	Port     string `env:"MATERIALS_SERVICE_POSTGRES_PORT"`
}

type Logger struct {
	Host string `env:"LOGGER_SERVICE_HOST"`
	Port string `env:"LOGGER_SERVICE_PORT"`
}

type Kafka struct {
	Host            string `env:"KAFKA_HOST"`
	Port            string `env:"KAFKA_PORT"`
	MaterialDeleted string `env:"MATERIALS_DELETE_MATERIAL"`
}

func MustLoad() *Config {
	cfg := &Config{}
	err := cleanenv.ReadEnv(cfg)
	if err != nil {
		log.Fatalf("failed to read env variables: %s", err)
	}
	return cfg
}
