package handlers

import (
	"github.com/faust8888/shortener/internal/app/service"
	"github.com/faust8888/shortener/internal/app/util"
	"io"
	"net/http"
)

const LocationHeader = "Location"

type Handler struct {
	URLShortener service.URLShortener
}

func (h *Handler) CreateShortURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "Couldn't read the targetFullURL of request!", http.StatusBadRequest)
		return
	}

	shortURL, err := h.URLShortener.CreateShortURL(string(body))

	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(shortURL))
}

func (h *Handler) GetFullURL(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := util.ExtractHashURLFromPath(req.URL.Path)
	fullURL, err := h.URLShortener.FindFullURL(searchedHashURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func CreateInMemoryHandler() *Handler {
	return &Handler{URLShortener: service.NewInMemoryShortenerService()}
}
