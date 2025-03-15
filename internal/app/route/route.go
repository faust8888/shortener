package route

import (
	"github.com/faust8888/shortener/internal/app/logger"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const HashKeyURLQueryParam = "hashKeyURL"

type route interface {
	CreateShortURL(res http.ResponseWriter, req *http.Request)
	GetFullURL(res http.ResponseWriter, req *http.Request)
}

func Create(r route) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Post("/", r.CreateShortURL)
	router.Get("/{"+HashKeyURLQueryParam+"}", r.GetFullURL)
	return router
}
