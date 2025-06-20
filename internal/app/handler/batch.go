package handler

import (
	"encoding/json"
	"fmt"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/security"
	"net/http"
)

type batch struct {
	service batchSaver
	authKey string
}

type batchSaver interface {
	CreateWithBatch(batch []model.CreateShortRequestBatchItemRequest, userID string) ([]model.CreateShortRequestBatchItemResponse, error)
}

func (handler *batch) CreateWithBatch(res http.ResponseWriter, req *http.Request) {
	token := security.GetToken(req)
	if token == "" {
		newToken, err := security.BuildToken(handler.authKey)
		if err != nil {
			http.Error(res, fmt.Sprintf("build token: %s", err.Error()), http.StatusInternalServerError)
			return
		}
		http.SetCookie(res, &http.Cookie{
			Name:  security.AuthorizationTokenName,
			Value: newToken,
		})
		token = newToken
	}
	userID, err := security.GetUserID(token, handler.authKey)
	if err != nil {
		http.Error(res, "unauthorized", http.StatusUnauthorized)
		return
	}

	decoder := json.NewDecoder(http.MaxBytesReader(nil, req.Body, 10<<20))
	var batchRequest []model.CreateShortRequestBatchItemRequest
	if err = decoder.Decode(&batchRequest); err != nil {
		http.Error(res, "Invalid request payload", http.StatusBadRequest)
		return
	}
	batchResponse, err := handler.service.CreateWithBatch(batchRequest, userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	resp, err := json.Marshal(&batchResponse)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
