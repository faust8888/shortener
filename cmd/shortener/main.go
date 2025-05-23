package main

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/handler"
	"github.com/faust8888/shortener/internal/app/migration"
	"github.com/faust8888/shortener/internal/app/repository"
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

	var repo repository.Repository
	if cfg.DataSourceName != "" {
		err := migration.Run(cfg.DataSourceName)
		if err != nil {
			logger.Log.Fatal("migration.run", zap.Error(err))
		}
		repo = postgres.NewPostgresRepository(cfg)
	} else {
		repo = inmemory.NewInMemoryRepository(cfg)
	}

	shortener := service.CreateShortener(repo, cfg.BaseShortURL)
	h := handler.Create(shortener, repo, cfg)
	logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, route.Create(h)); err != nil {
		panic(err)
	}
}
