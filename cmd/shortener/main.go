// Package main contains the entry point for the shortener service.
//
// This application provides a URL shortening service that can use either an in-memory store or PostgreSQL backend.
// Build information (version, date, commit) is injected at compile time and logged on startup.
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
	_ "net/http/pprof" // Import pprof for profiling endpoints
)

// BuildVersion is the semantic version of the build (e.g., v1.0.0).
// This value is typically injected at compile time via -ldflags.
var BuildVersion string = "N/A"

// BuildDate is the timestamp when the binary was built, usually in ISO format.
// This value is typically injected at compile time via -ldflags.
var BuildDate string = "N/A"

// BuildCommit is the git commit hash used to build the binary.
// This value is typically injected at compile time via -ldflags.
var BuildCommit string = "N/A"

// main starts the shortener service.
//
// It initializes logging, configures the repository based on configuration,
// sets up routing, and starts the HTTP server.
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
	h := handler.CreateHandler(shortener, repo, cfg)

	// Log build metadata
	logger.Log.Info("Build Info",
		zap.String("version", BuildVersion),
		zap.String("date", BuildDate),
		zap.String("commit", BuildCommit),
	)

	logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddress))
	if err := http.ListenAndServe(cfg.ServerAddress, route.Create(h)); err != nil {
		panic(err)
	}
}
