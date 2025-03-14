package main

import (
	"github.com/faust8888/shortener/cmd/config"
	"github.com/faust8888/shortener/internal/app/handlers"
	"github.com/faust8888/shortener/internal/app/logger"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	config.LoadConfig()
	if err := logger.Initialize(config.Cfg.LoggingLevel); err != nil {
		panic(err)
	}
	router := handlers.CreateRouter(handlers.CreateInMemoryHandler())
	logger.Log.Info("Starting server", zap.String("address", config.Cfg.ServerAddress))

	if err := http.ListenAndServe(config.Cfg.ServerAddress, router); err != nil {
		panic(err)
	}
}
