package route

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/middleware/compress"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
	"net/http/pprof"
)

type route interface {
	CreateWithBatch(res http.ResponseWriter, req *http.Request)
	CreateWithJSON(res http.ResponseWriter, req *http.Request)
	Create(res http.ResponseWriter, req *http.Request)
	FindByHash(res http.ResponseWriter, req *http.Request)
	FindByUserID(res http.ResponseWriter, req *http.Request)
	Ping(res http.ResponseWriter, req *http.Request)
	Delete(res http.ResponseWriter, req *http.Request)
}

// Create инициализирует HTTP-роутер на основе chi и регистрирует маршруты,
// связывая каждый маршрут с соответствующим методом интерфейса route.
//
// Поддерживаемые маршруты:
// - POST /api/shorten          → CreateWithJSON
// - POST /api/shorten/batch    → CreateWithBatch
// - POST /                     → Create
// - GET /{hash}                → FindByHash
// - GET /api/user/urls         → FindByUserID
// - GET /ping                  → Ping
// - DELETE /api/user/urls      → Delete
// - /debug/pprof/*             → pprof (для профилирования)
//
// Возвращает:
//   - *chi.Mux: готовый к использованию HTTP-роутер.
func Create(r route) *chi.Mux {
	router := chi.NewRouter()
	router.Use(gzip.NewMiddleware)
	router.Use(logger.NewMiddleware)
	router.Post("/api/shorten", r.CreateWithJSON)
	router.Post("/api/shorten/batch", r.CreateWithBatch)
	router.Post("/", r.Create)
	router.Get("/{"+config.HashKeyURLQueryParam+"}", r.FindByHash)
	router.Get("/api/user/urls", r.FindByUserID)
	router.Get("/ping", r.Ping)
	router.Delete("/api/user/urls", r.Delete)
	router.Get("/debug/pprof/*", pprof.Index)
	router.Get("/debug/pprof/cmdline", pprof.Cmdline)
	router.Get("/debug/pprof/profile", pprof.Profile)
	router.Get("/debug/pprof/symbol", pprof.Symbol)
	router.Get("/debug/pprof/trace", pprof.Trace)
	return router
}
