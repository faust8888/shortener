// Package main contains the entry point for the shortener service.
//
// This application provides a URL shortening service that can use either an in-memory store or PostgreSQL backend.
// Build information (version, date, commit) is injected at compile time and logged on startup.
package main

import (
	"context"
	"errors"
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
	"golang.org/x/sync/errgroup"
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
// sets up routing, and starts the HTTP server with graceful shutdown.
func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Application error: %v\n", err)
		os.Exit(1)
	}
}

// run contains the main application logic and returns an error.
func run() error {
	cfg := config.Create()
	if err := logger.Initialize(cfg.LoggingLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize repository
	var repo repository.Repository
	if cfg.DataSourceName != "" {
		err := migration.Run(cfg.DataSourceName)
		if err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
		repo = postgres.NewPostgresRepository(cfg)
	} else {
		repo = inmemory.NewInMemoryRepository(cfg)
	}

	// Ensure repository cleanup
	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Log.Error("Failed to close repository", zap.Error(err))
			}
		}
	}()

	shortener := service.CreateShortener(repo, cfg.BaseShortURL)
	h := handler.CreateHandler(shortener, repo, cfg)

	// Log build metadata
	printBuildInfo()

	// --- Graceful Shutdown with errgroup ---
	// Create a context that is canceled when a termination signal is received.
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	defer stop()

	// Create an errgroup with the context.
	g, gctx := errgroup.WithContext(ctx)

	// Create the HTTP server.
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: route.Create(h),
	}

	// Goroutine to run the HTTP server.
	g.Go(func() error {
		logger.Log.Info("Starting server",
			zap.String("address", cfg.ServerAddress),
			zap.Bool("https", cfg.EnableHTTPS))

		var err error
		if cfg.EnableHTTPS {
			manager := &autocert.Manager{
				Cache:      autocert.DirCache("cache-dir"),
				Prompt:     autocert.AcceptTOS,
				HostPolicy: autocert.HostWhitelist("localhost", "127.0.0.1"),
			}
			server.TLSConfig = manager.TLSConfig()
			err = server.ListenAndServeTLS("", "")
		} else {
			err = server.ListenAndServe()
		}

		// http.ErrServerClosed is the expected error during graceful shutdown.
		if errors.Is(err, http.ErrServerClosed) {
			logger.Log.Info("Server stopped")
			return nil
		}
		return fmt.Errorf("server failed: %w", err)
	})

	// Goroutine to handle graceful shutdown.
	g.Go(func() error {
		// Wait for the context to be canceled (i.e., a signal is received).
		<-gctx.Done()

		logger.Log.Info("Shutting down server gracefully...")

		// Create a separate context with a timeout for the shutdown process.
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Perform the graceful shutdown.
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}

		logger.Log.Info("Server gracefully stopped")
		return nil
	})

	// Wait for all goroutines in the group to complete.
	if err := g.Wait(); err != nil {
		// Filter out the expected context cancellation error.
		if errors.Is(err, context.Canceled) {
			logger.Log.Info("Application stopped successfully")
			return nil
		}
		return fmt.Errorf("application error: %w", err)
	}

	logger.Log.Info("Application stopped successfully")
	return nil
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
