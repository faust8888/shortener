package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
)

const (
	ServerAddressFlag            = "a"
	BaseShortURLFlag             = "b"
	ServerAddressEnvironmentName = "SERVER_ADDRESS"
	BaseShortURLEnvironmentName  = "BASE_URL"
)

var Cfg Config

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseShortURL  string `env:"BASE_URL"`
}

func LoadConfig() {
	if Cfg.ServerAddress == "" {
		flag.StringVar(&Cfg.ServerAddress, ServerAddressFlag, "localhost:8080", "Address of the running server")
	}
	if Cfg.BaseShortURL == "" {
		flag.StringVar(&Cfg.BaseShortURL, BaseShortURLFlag, "http://localhost:8080", "Base URL for returning short URL")
	}
	flag.Parse()
	err := env.Parse(&Cfg)
	if err != nil {
		fmt.Printf("couldn't parse environment variables: %s\n", err.Error())
	}
}
