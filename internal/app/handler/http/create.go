package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// Create — это HTTP-обработчик для создания короткой ссылки из обычного текста.
// Использует интерфейс creator для сохранения данных и требует ключ аутентификации.
type Create struct {
	service creator
	authKey string
}

type creator interface {
	Create(fullURL string, userID string) (string, error)
}

// CreateLink обрабатывает POST-запрос на создание короткой ссылки.
//
// Метод:
// - Читает тело запроса как plain text.
// - Проверяет или генерирует токен авторизации.
// - Извлекает userID из токена.
// - Передаёт данные сервису для сохранения.
// - Возвращает короткий URL в теле ответа.
//
// Возможные HTTP-статусы:
// - 201 Created — успешно создано.
// - 400 Bad Request — невалидное тело запроса.
// - 401 Unauthorized — отсутствующий или недействительный токен.
// - 409 Conflict — дублирующаяся запись.
// - 500 Internal Server Error — внутренняя ошибка сервера.
func (handler *Create) CreateLink(res http.ResponseWriter, req *http.Request) {
	requestBody, err := io.ReadAll(req.Body)
	if err != nil {
		http.Error(res, "couldn't read the targetFullURL of request!", http.StatusBadRequest)
		return
	}

	token := security.GetToken(req)
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
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}

	fullURL := string(requestBody)
	shortURL, err := handler.service.Create(fullURL, userID)
	isUniqueConstraintViolation := errors.Is(err, postgres.ErrUniqueIndexConstraint)
	if err != nil && !isUniqueConstraintViolation {
		logger.Log.Error("Failed to CreateLink short URL", zap.String("body", fullURL), zap.Error(err))
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

// CreateWithJSON — HTTP-обработчик для создания короткой ссылки через JSON.
// Использует тот же интерфейс creator для сохранения данных.
type CreateWithJSON struct {
	service creator
	authKey string
}

// CreateLinkWithJSON обрабатывает POST-запрос с JSON-телом вида {"url": "http://example.com"}.
//
// Метод:
// - Читает и парсит JSON-запрос.
// - Проверяет или генерирует токен авторизации.
// - Извлекает userID из токена.
// - Передаёт данные сервису для сохранения.
// - Возвращает JSON-ответ с результатом.
//
// Пример тела запроса:
//
//	{"url": "http://example.com"}
//
// Ответ:
//
//	{"result": "http://your-shortener.com/abc"}
//
// Возможные HTTP-статусы:
// - 201 Created — успешно создано.
// - 400 Bad Request — невалидное тело запроса.
// - 401 Unauthorized — отсутствующий или недействительный токен.
// - 409 Conflict — дублирующаяся запись.
// - 500 Internal Server Error — внутренняя ошибка сервера.
func (handler *CreateWithJSON) CreateLinkWithJSON(res http.ResponseWriter, req *http.Request) {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}

	token := security.GetToken(req)
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
		http.Error(res, err.Error(), http.StatusUnauthorized)
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

	shortURL, err := handler.service.Create(createRequest.URL, userID)
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
