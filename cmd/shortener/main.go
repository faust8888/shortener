// Package main contains the entry point for the shortener service.
//
// This application provides a URL shortening service that can use either an in-memory store or PostgreSQL backend.
// Build information (version, date, commit) is injected at compile time and logged on startup.
package main

import (
	"context"
	"fmt"
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
	"golang.org/x/crypto/acme/autocert"
	"net/http"
	_ "net/http/pprof" // Import pprof for profiling endpoints
	"os"
	"os/signal"
	"syscall"
	"time"
)

// buildVersion is the semantic version of the build (e.g., v1.0.0).
// This value is typically injected at compile time via -ldflags.
var buildVersion string = "N/A"

// buildDate is the timestamp when the binary was built, usually in ISO format.
// This value is typically injected at compile time via -ldflags.
var buildDate string = "N/A"

// buildCommit is the git commit hash used to build the binary.
// This value is typically injected at compile time via -ldflags.
var buildCommit string = "N/A"

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
	printBuildInfo()

	// --- Настройка сервера и Graceful Shutdown ---
	// Создаем канал для получения сигналов ОС.
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Создаем экземпляр http.Server для полного контроля.
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: route.Create(h),
	}

	// Запускаем сервер в отдельной горутине.
	go func() {
		logger.Log.Info("Starting server", zap.String("address", cfg.ServerAddress), zap.Bool("https", cfg.EnableHTTPS))
		var err error
		if cfg.EnableHTTPS {
			manager := &autocert.Manager{
				Cache:      autocert.DirCache("cache-dir"),
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist("localhost", "127.0.0.1"), // Укажите ваши домены
			}
			server.TLSConfig = manager.TLSConfig()
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}

		// http.ErrServerClosed - ожидаемая ошибка при вызове Shutdown, ее не логируем как фатальную.
		if err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Блокируем основной поток до получения сигнала.
	<-sigChan

	logger.Log.Info("Shutting down server gracefully...")

	// Создаем контекст с таймаутом для завершения работы.
	// Даем серверу 10 секунд на завершение обработки текущих запросов.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Вызываем Shutdown, который мягко завершает работу сервера.
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
	}

	logger.Log.Info("Server gracefully stopped")
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
