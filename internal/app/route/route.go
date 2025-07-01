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
	CreateLinkWithBatch(res http.ResponseWriter, req *http.Request)
	CreateLinkWithJSON(res http.ResponseWriter, req *http.Request)
	CreateLink(res http.ResponseWriter, req *http.Request)
	FindLinkByHash(res http.ResponseWriter, req *http.Request)
	FindLinkByUserID(res http.ResponseWriter, req *http.Request)
	PingDatabase(res http.ResponseWriter, req *http.Request)
	DeleteLink(res http.ResponseWriter, req *http.Request)
}

// Create инициализирует HTTP-роутер на основе chi и регистрирует маршруты,
// связывая каждый маршрут с соответствующим методом интерфейса route.
//
// Поддерживаемые маршруты:
// - POST /api/shorten          → CreateLinkWithJSON
// - POST /api/shorten/batch    → CreateLinkWithBatch
// - POST /                     → Create
// - GET /{hash}                → FindLinkByHash
// - GET /api/user/urls         → FindLinkByUserID
// - GET /ping                  → PingDatabase
// - DELETE /api/user/urls      → DeleteLink
// - /debug/pprof/*             → pprof (для профилирования)
//
// Возвращает:
//   - *chi.Mux: готовый к использованию HTTP-роутер.
func Create(r route) *chi.Mux {
	router := chi.NewRouter()
	router.Use(gzip.NewMiddleware)
	router.Use(logger.NewMiddleware)
	router.Post("/api/shorten", r.CreateLinkWithJSON)
	router.Post("/api/shorten/batch", r.CreateLinkWithBatch)
	router.Post("/", r.CreateLink)
	router.Get("/{"+config.HashKeyURLQueryParam+"}", r.FindLinkByHash)
	router.Get("/api/user/urls", r.FindLinkByUserID)
	router.Get("/ping", r.PingDatabase)
	router.Delete("/api/user/urls", r.DeleteLink)
	router.Get("/debug/pprof/*", pprof.Index)
	router.Get("/debug/pprof/cmdline", pprof.Cmdline)
	router.Get("/debug/pprof/profile", pprof.Profile)
	router.Get("/debug/pprof/symbol", pprof.Symbol)
	router.Get("/debug/pprof/trace", pprof.Trace)
	return router
}
