package handlers

import (
	"fmt"
	"github.com/faust8888/shortener/internal/app/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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
	fmt.Printf("\ncreating short URL for '%s'\n", fullURL)
	shortURL, err := h.URLShortener.CreateShortURL(fullURL)

	if err != nil {
		fmt.Printf("couldn't create: '%s'!\n", err.Error())
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("created: '%s' -> '%s'\n", fullURL, shortURL)

	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		fmt.Printf("couldn't write response: '%s'\n", err.Error())
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) GetFullURL(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := chi.URLParam(req, HashKeyURLQueryParam)
	fmt.Printf("\nfinding short URL by '%s' \n", searchedHashURL)

	fullURL, err := h.URLShortener.FindFullURL(searchedHashURL)
	if err != nil {
		fmt.Printf("not found: '%s'\n", err.Error())
		http.Error(res, err.Error(), http.StatusNotFound)
	}

	fmt.Printf("found: '%s'\n", fullURL)
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func CreateInMemoryHandler() *Handler {
	fmt.Println("Creating in memory handler")
	return &Handler{URLShortener: service.NewInMemoryShortenerService()}
}

func CreateRouter(h *Handler) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Post("/", h.CreateShortURL)
	router.Get("/{"+HashKeyURLQueryParam+"}", h.GetFullURL)
	return router
}
