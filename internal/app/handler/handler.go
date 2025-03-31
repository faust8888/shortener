package handler

import (
	"github.com/faust8888/shortener/internal/app/service"
)

type Handler struct {
	create
	createWithJSON
	find
	ping
}

func Create(s *service.Shortener, pingChecker PingChecker) *Handler {
	return &Handler{
		create:         create{service: s},
		createWithJSON: createWithJSON{s},
		find:           find{s},
		ping:           ping{service: pingChecker},
	}
}
