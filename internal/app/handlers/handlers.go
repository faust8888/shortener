package handlers

import (
	"github.com/faust8888/shortener/internal/app/logger"
	"github.com/faust8888/shortener/internal/app/service"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
)

const (
	LocationHeader       = "Location"
	HashKeyURLQueryParam = "hashKeyURL"
)

type Handler struct {
	URLShortener service.URLShortener
}

func (h *Handler) CreateShortURL(res http.ResponseWriter, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "couldn't read the targetFullURL of request!", http.StatusBadRequest)
		return
	}

	fullURL := string(requestBody)
	shortURL, err := h.URLShortener.CreateShortURL(fullURL)

	if err != nil {
		logger.Log.Error("Failed to create short URL", zap.String("url", fullURL), zap.Error(err))
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Log.Info("created short URL", zap.String("shortUrl", shortURL), zap.String("fullUrl", fullURL))
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		logger.Log.Error("couldn't write response", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetFullURL(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := chi.URLParam(req, HashKeyURLQueryParam)
	logger.Log.Info("getting full URL", zap.String("searchedHashURL", searchedHashURL))
	fullURL, err := h.URLShortener.FindFullURL(searchedHashURL)
	if err != nil {
		logger.Log.Error("couldn't find short URL", zap.Error(err))
		http.Error(res, err.Error(), http.StatusNotFound)
	}
	logger.Log.Info("found short URL", zap.String("url", fullURL))
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func CreateInMemoryHandler() *Handler {
	logger.Log.Info("Creating in memory handler")
	return &Handler{URLShortener: service.NewInMemoryShortenerService()}
}

func CreateRouter(h *Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(logger.RequestLogger)
	router.Post("/", h.CreateShortURL)
	router.Get("/{"+HashKeyURLQueryParam+"}", h.GetFullURL)
	return router
}
