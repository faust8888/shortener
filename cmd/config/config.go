package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/faust8888/shortener/internal/app/logger"
	"go.uber.org/zap"
)

const (
	ServerAddressFlag            = "a"
	BaseShortURLFlag             = "b"
	LoggingLevelFlag             = "l"
	ServerAddressEnvironmentName = "SERVER_ADDRESS"
	BaseShortURLEnvironmentName  = "BASE_URL"
)

var Cfg Config

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseShortURL  string `env:"BASE_URL"`
	LoggingLevel  string `env:"LOGGING_LEVEL"`
}

func LoadConfig() {
	if Cfg.ServerAddress == "" {
		flag.StringVar(&Cfg.ServerAddress, ServerAddressFlag, "localhost:8080", "Address of the running server")
	}
	if Cfg.BaseShortURL == "" {
		flag.StringVar(&Cfg.BaseShortURL, BaseShortURLFlag, "http://localhost:8080", "Base URL for returning short URL")
	}
	if Cfg.LoggingLevel == "" {
		flag.StringVar(&Cfg.LoggingLevel, LoggingLevelFlag, "INFO", "Level of logging to use")
	}
	flag.Parse()
	err := env.Parse(&Cfg)
	if err != nil {
		logger.Log.Error("Failed to parse config", zap.Error(err))
	}
}
