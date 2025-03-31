package route

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/middleware/compress"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type route interface {
	CreateWithJSON(res http.ResponseWriter, req *http.Request)
	Create(res http.ResponseWriter, req *http.Request)
	Get(res http.ResponseWriter, req *http.Request)
	Ping(res http.ResponseWriter, req *http.Request)
}

func Create(r route) *chi.Mux {
	router := chi.NewRouter()
	router.Use(gzip.NewMiddleware)
	router.Use(logger.NewMiddleware)
	router.Post("/api/shorten", r.CreateWithJSON)
	router.Post("/", r.Create)
	router.Get("/{"+config.HashKeyURLQueryParam+"}", r.Get)
	router.Get("/ping", r.Ping)
	return router
}
