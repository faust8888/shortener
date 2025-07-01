package logger

import (
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Log — глобальный логгер приложения.
// По умолчанию установлен в "no-op" режим до инициализации.
var Log = zap.NewNop()

// Initialize настраивает глобальный логгер на основе указанного уровня логирования.
//
// Параметры:
//   - level: строковое представление уровня логирования (например, "debug", "info").
//
// Возвращает:
//   - error: nil, если инициализация прошла успешно, иначе — ошибку.
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

// NewMiddleware возвращает HTTP middleware, который логирует входящие запросы и исходящие ответы.
//
// Для каждого запроса логируются:
//   - HTTP-метод,
//   - URL-путь,
//   - статус ответа,
//   - объём переданных данных,
//   - время выполнения.
func NewMiddleware(handler http.Handler) http.Handler {
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
