package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
)

type create struct {
	service creator
}

type creator interface {
	Create(fullURL string) (string, error)
}

func (handler *create) Create(res http.ResponseWriter, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "couldn't read the targetFullURL of request!", http.StatusBadRequest)
		return
	}

	fullURL := string(requestBody)
	shortURL, err := handler.service.Create(fullURL)
	isUniqueConstraintViolation := errors.Is(err, postgres.ErrUniqueIndexConstraint)
	if err != nil && !isUniqueConstraintViolation {
		logger.Log.Error("Failed to create short URL", zap.String("body", fullURL), zap.Error(err))
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	if isUniqueConstraintViolation {
		res.WriteHeader(http.StatusConflict)
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	_, err = res.Write([]byte(shortURL))
	if err != nil {
		logger.Log.Error("couldn't write response", zap.Error(err))
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

type createWithJSON struct {
	service creator
}

func (handler *createWithJSON) CreateWithJSON(res http.ResponseWriter, req *http.Request) {
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

	shortURL, err := handler.service.Create(createRequest.URL)
	isUniqueConstraintViolation := errors.Is(err, postgres.ErrUniqueIndexConstraint)
	if err != nil && !isUniqueConstraintViolation {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := json.Marshal(&model.CreateShortResponse{Result: shortURL})
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	if isUniqueConstraintViolation {
		res.WriteHeader(http.StatusConflict)
	} else {
		res.WriteHeader(http.StatusCreated)
	}

	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
