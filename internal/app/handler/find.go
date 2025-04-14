package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const LocationHeader = "Location"

type find struct {
	service finder
}

type finder interface {
	Find(hashURL string) (string, error)
}

func (handler *find) Get(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := chi.URLParam(req, config.HashKeyURLQueryParam)
	fullURL, err := handler.service.Find(searchedHashURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
