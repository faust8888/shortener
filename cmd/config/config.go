package config

import (
	"flag"
)

const (
	ServerAddressFlag = "a"
	BaseShortURLFlag  = "b"
)

var Config struct {
	ServerAddress string
	BaseShortURL  string
}

func init() {
	flag.StringVar(&Config.ServerAddress, ServerAddressFlag, "localhost:8080", "Address of the running server")
	flag.StringVar(&Config.BaseShortURL, BaseShortURLFlag, "http://localhost:8080", "Base URL for returning short URL")
}
