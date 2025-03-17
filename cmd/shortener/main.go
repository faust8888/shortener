package main

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/handler"
	"github.com/faust8888/shortener/internal/app/logger"
	"github.com/faust8888/shortener/internal/app/repository/inmemory"
	"github.com/faust8888/shortener/internal/app/route"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	cfg := config.Create()
	if err := logger.Initialize(cfg.LoggingLevel); err != nil {
		panic(err)
	}
	h := handler.Create(inmemory.NewRepository(), cfg.BaseShortURL)
	logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, route.Create(h)); err != nil {
		panic(err)
	}
}
