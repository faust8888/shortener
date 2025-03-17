package handler

import (
	"bytes"
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/logger"
	"github.com/faust8888/shortener/internal/app/model"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type post struct {
	creator creator
}

type creator interface {
	Create(fullURL string) (string, error)
}

func (handler *post) Create(res http.ResponseWriter, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "couldn't read the targetFullURL of request!", http.StatusBadRequest)
		return
	}

	fullURL := string(requestBody)
	shortURL, err := handler.creator.Create(fullURL)

	if err != nil {
		logger.Log.Error("Failed to post short URL", zap.String("body", fullURL), zap.Error(err))
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Log.Info("created short URL", zap.String("shortUrl", shortURL), zap.String("fullUrl", fullURL))
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write([]byte(shortURL))
	if err != nil {
		logger.Log.Error("couldn't write response", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

type postWithJSON struct {
	creator creator
}

func (handler *postWithJSON) CreateWithJSON(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	var createRequest model.CreateShortRequest
	if err = json.Unmarshal(buf.Bytes(), &createRequest); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if err = createRequest.Validate(); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	shortURL, _ := handler.creator.Create(createRequest.URL)
	resp, err := json.Marshal(&model.CreateShortResponse{Result: shortURL})
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusCreated)
	_, err = res.Write(resp)
	if err != nil {
		logger.Log.Error("couldn't write response", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
