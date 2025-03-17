package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
)

const (
	LocationHeader = "Location"
)

type get struct {
	finder finder
}

type finder interface {
	Find(hashURL string) (string, error)
}

func (handler *get) Get(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := chi.URLParam(req, config.HashKeyURLQueryParam)
	logger.Log.Info("getting full URL", zap.String("searchedHashURL", searchedHashURL))
	fullURL, err := handler.finder.Find(searchedHashURL)
	if err != nil {
		logger.Log.Error("couldn't find short URL", zap.Error(err))
		http.Error(res, err.Error(), http.StatusNotFound)
	}
	logger.Log.Info("found short URL", zap.String("body", fullURL))
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}
