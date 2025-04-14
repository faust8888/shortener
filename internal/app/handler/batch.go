package handler

import (
	"bytes"
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
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	token := security.GetToken(req.Cookies())
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
	userID, err := security.GetUserID(token, handler.authKey)
	if err != nil {
		http.Error(res, "unauthorized", http.StatusUnauthorized)
		return
	}

	var batchRequest []model.CreateShortRequestBatchItemRequest
	if err = json.Unmarshal(buf.Bytes(), &batchRequest); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
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
