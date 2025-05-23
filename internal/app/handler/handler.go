package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/service"
)

type Handler struct {
	create
	createWithJSON
	batch
	find
	ping
	delete
}

func Create(s *service.Shortener, pingChecker PingChecker, cfg *config.Config) *Handler {
	return &Handler{
		create:         create{service: s, authKey: cfg.AuthKey},
		createWithJSON: createWithJSON{service: s, authKey: cfg.AuthKey},
		batch:          batch{service: s, authKey: cfg.AuthKey},
		find:           find{service: s, authKey: cfg.AuthKey},
		ping:           ping{pingChecker},
		delete:         delete{service: s, authKey: cfg.AuthKey},
	}
}
