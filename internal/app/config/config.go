package config

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

const (
	ServerAddressFlag    = "a"
	BaseShortURLFlag     = "b"
	LoggingLevelFlag     = "l"
	StorageFilePathFlag  = "f"
	HashKeyURLQueryParam = "hashKeyURL"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseShortURL    string `env:"BASE_URL"`
	LoggingLevel    string `env:"LOGGING_LEVEL"`
	StorageFilePath string `env:"FILE_STORAGE_PATH"`
}

func Create() *Config {
	var cfg Config

	setFlag(&cfg.ServerAddress, ServerAddressFlag, "localhost:8080", "Address of the server")
	setFlag(&cfg.BaseShortURL, BaseShortURLFlag, "http://localhost:8080", "Base URL for returning short URL")
	setFlag(&cfg.LoggingLevel, LoggingLevelFlag, "INFO", "Level of logging to use")
	setFlag(&cfg.StorageFilePath, StorageFilePathFlag, "./storage.txt", "Path to the storage file")
	flag.Parse()

	err := env.Parse(&cfg)
	if err != nil {
		logger.Log.Error("Failed to parse config", zap.Error(err))
	}

	return &cfg
}

func setFlag(p *string, flagName string, defaultFlagValue string, description string) {
	if flag.Lookup(flagName) == nil {
		flag.StringVar(p, flagName, defaultFlagValue, description)
	} else {
		*p = flag.Lookup(flagName).Value.String()
	}
}
