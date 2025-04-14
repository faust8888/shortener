package handler

import (
	"net/http"
)

type ping struct {
	service PingChecker
}

type PingChecker interface {
	Ping() (bool, error)
}

func (handler *ping) Ping(res http.ResponseWriter, req *http.Request) {
	_, err := handler.service.Ping()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}
