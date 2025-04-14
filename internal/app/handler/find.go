package handler

import (
	"encoding/json"
	"fmt"
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const LocationHeader = "Location"

type find struct {
	service finder
	authKey string
}

type finder interface {
	FindByHash(hashURL string) (string, error)
	FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error)
}

func (handler *find) FindByHash(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := chi.URLParam(req, config.HashKeyURLQueryParam)
	fullURL, err := handler.service.FindByHash(searchedHashURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func (handler *find) FindByUserID(res http.ResponseWriter, req *http.Request) {
	token := security.GetToken(req.Cookies())
	userID, err := security.GetUserID(token, handler.authKey)
	if token == "" {
		token, err = security.BuildToken(handler.authKey)
		if err != nil {
			http.Error(res, fmt.Sprintf("build token: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:  security.AuthorizationTokenName,
			Value: token,
		})
	}
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}
	fullURL, err := handler.service.FindAllByUserID(userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(fullURL) == 0 {
		http.Error(res, "no content", http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(&fullURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
