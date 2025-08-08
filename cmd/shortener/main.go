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
	http2 "github.com/faust8888/shortener/internal/app/handler/http"
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
	"google.golang.org/grpc"
	"net"
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

	defer func() {
		if closer, ok := repo.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				logger.Log.Error("Failed to close repository", zap.Error(err))
			}
		}
	}()

	printBuildInfo()
	shortener := service.CreateShortener(repo, cfg.BaseShortURL)

	// --- Setup Signal Handling ---
	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	defer stop()

	// Use errgroup to manage both servers
	g, gctx := errgroup.WithContext(ctx)

	// --- Start HTTP Server ---
	g.Go(func() error {
		err, _ := runHTTPWithContext(gctx, shortener, repo, cfg)
		return err
	})

	// --- Start gRPC Server ---
	if cfg.EnableGRPC {
		g.Go(func() error {
			return runGRPCWithContext(gctx, cfg)
		})
	}

	// Wait for either server to fail or signal to terminate
	if err := g.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("application error: %w", err)
	}

	logger.Log.Info("Application stopped successfully")
	return nil
}

func runHTTPWithContext(ctx context.Context, shortener *service.Shortener, repo repository.Repository, cfg *config.Config) (error, bool) {
	h := http2.CreateHandler(shortener, repo, cfg)

	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: route.Create(h),
	}

	// Goroutine to start server
	go func() {
		logger.Log.Info("Starting HTTP server",
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

		if !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("HTTP server failed", zap.Error(err))
		}
	}()

	// Goroutine to handle shutdown
	go func() {
		<-ctx.Done()
		logger.Log.Info("Shutting down HTTP server gracefully...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error("HTTP shutdown failed", zap.Error(err))
		}
	}()

	// Block until context is done
	<-ctx.Done()
	return nil, true
}

func runGRPCWithContext(ctx context.Context, cfg *config.Config) error {
	s := grpc.NewServer()
	lis, err := net.Listen("tcp", cfg.ServerGRPCAddress)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", cfg.ServerGRPCAddress, err)
	}

	// Serve gRPC in a goroutine
	go func() {
		logger.Log.Info("Starting gRPC server", zap.String("address", cfg.ServerGRPCAddress))
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logger.Log.Error("gRPC server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	logger.Log.Info("Shutting down gRPC server gracefully...")
	shutdownDone := make(chan struct{})
	go func() {
		s.GracefulStop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		return nil
	case <-time.After(10 * time.Second):
		s.Stop() // Force stop after timeout
		return context.DeadlineExceeded
	}
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
