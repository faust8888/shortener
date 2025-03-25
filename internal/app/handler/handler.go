package handler

import (
	"github.com/faust8888/shortener/internal/app/repository"
	"github.com/faust8888/shortener/internal/app/service"
)

type Handler struct {
	post
	postWithJSON
	get
}

func Create(s repository.Repository, baseShortURL string) *Handler {
	return &Handler{
		post:         post{service: service.CreateShortener(s, baseShortURL)},
		postWithJSON: postWithJSON{service: service.CreateShortener(s, baseShortURL)},
		get:          get{service: service.CreateShortener(s, baseShortURL)},
	}
}
