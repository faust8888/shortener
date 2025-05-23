package handler

import (
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/security"
	"io"
	"net/http"
)

type delete struct {
	service deleter
	authKey string
}

type deleter interface {
	DeleteAsync(ids []string, userID string) error
}

func (handler *delete) Delete(res http.ResponseWriter, req *http.Request) {
	token := security.GetToken(req)
	if token == "" {
		res.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, err := security.GetUserID(token, handler.authKey)
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "couldn't read the targetFullURL of request!", http.StatusBadRequest)
		return
	}

	var ids []string
	err = json.Unmarshal(requestBody, &ids)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	err = handler.service.DeleteAsync(ids, userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	res.WriteHeader(http.StatusAccepted)
}
