package handler

import (
	"bytes"
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/model"
	"net/http"
)

type batch struct {
	service batchSaver
}

type batchSaver interface {
	CreateWithBatch(batch []model.CreateShortRequestBatchItemRequest) ([]model.CreateShortRequestBatchItemResponse, error)
}

func (handler *batch) CreateWithBatch(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var batchRequest []model.CreateShortRequestBatchItemRequest
	if err = json.Unmarshal(buf.Bytes(), &batchRequest); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	batchResponse, err := handler.service.CreateWithBatch(batchRequest)
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
