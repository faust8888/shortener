package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var Log *zap.Logger = zap.NewNop()

func Initialize(level string) error {
	loggingLevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	cfg := zap.NewProductionConfig()
	cfg.Level = loggingLevel
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	Log = zl
	return nil
}

func RequestLogger(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrappedWriter := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		Log.Info("----> incoming HTTP request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
		)
		handler.ServeHTTP(wrappedWriter, r)
		Log.Info("<---- outgoing HTTP response",
			zap.Int("status", wrappedWriter.Status()),
			zap.Int("size", wrappedWriter.BytesWritten()),
			zap.String("execution time", time.Since(start).String()),
		)
	})
}
