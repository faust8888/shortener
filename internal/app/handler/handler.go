package handler

import (
	"github.com/faust8888/shortener/internal/app/service"
)

type Handler struct {
	create
	createWithJSON
	batch
	find
	ping
}

func Create(s *service.Shortener, pingChecker PingChecker) *Handler {
	return &Handler{
		create:         create{s},
		createWithJSON: createWithJSON{s},
		batch:          batch{s},
		find:           find{s},
		ping:           ping{pingChecker},
	}
}
