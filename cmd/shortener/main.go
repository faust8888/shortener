package main

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/handler"
	"github.com/faust8888/shortener/internal/app/repository/inmemory"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/app/route"
	"github.com/faust8888/shortener/internal/app/service"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
	"net/http"
)

func main() {
	cfg := config.Create()
	if err := logger.Initialize(cfg.LoggingLevel); err != nil {
		panic(err)
	}

	shortener := service.CreateShortener(inmemory.NewInMemoryRepository(cfg.StorageFilePath), cfg.BaseShortURL)
	repo := postgres.NewPostgresRepository(cfg.DataSourceName)
	h := handler.Create(shortener, repo)

	logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, route.Create(h)); err != nil {
		panic(err)
	}
}
