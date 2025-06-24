package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/go-chi/chi/v5"
	"net/http"
)

const LocationHeader = "Location"

type find struct {
	service finder
	authKey string
}

type finder interface {
	FindByHash(hashURL string) (string, error)
	FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error)
}

// FindByHash обрабатывает GET-запрос на редирект по короткой ссылке.
//
// Метод:
// - Извлекает хэш из пути запроса.
// - Передаёт хэш сервису для поиска оригинального URL.
// - Возвращает редирект или соответствующую ошибку.
//
// Путь: /{hash}
//
// Возможные HTTP-статусы:
// - 307 Temporary Redirect — успешный редирект.
// - 404 Not Found — ссылка не найдена.
// - 410 Gone — ссылка была удалена.
func (handler *find) FindByHash(res http.ResponseWriter, req *http.Request) {
	searchedHashURL := chi.URLParam(req, config.HashKeyURLQueryParam)
	fullURL, err := handler.service.FindByHash(searchedHashURL)
	if errors.Is(err, postgres.ErrRecordWasMarkedAsDeleted) {
		res.WriteHeader(http.StatusGone)
		return
	}
	if err != nil {
		http.Error(res, err.Error(), http.StatusNotFound)
	}
	res.Header().Set(LocationHeader, fullURL)
	res.WriteHeader(http.StatusTemporaryRedirect)
}

// FindByUserID обрабатывает GET-запрос для получения всех сокращённых ссылок текущего пользователя.
//
// Метод:
// - Проверяет или генерирует токен авторизации.
// - Извлекает идентификатор пользователя из токена.
// - Передаёт запрос сервису для получения списка ссылок.
// - Возвращает JSON-ответ со списком ссылок.
//
// Пример ответа:
//
//	[
//	  {"short_url": "http://localhost:8080/abc", "original_url": "http://example.com"},
//	  {"short_url": "http://localhost:8080/def", "original_url": "http://example.org"}
//	]
//
// Возможные HTTP-статусы:
// - 200 OK — успешно возвращён список ссылок.
// - 204 No Content — у пользователя нет сохранённых ссылок.
// - 401 Unauthorized — отсутствующий или недействительный токен.
// - 500 Internal Server Error — внутренняя ошибка сервера.
func (handler *find) FindByUserID(res http.ResponseWriter, req *http.Request) {
	token := security.GetToken(req)
	userID, err := security.GetUserID(token, handler.authKey)
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
	if err != nil {
		http.Error(res, err.Error(), http.StatusUnauthorized)
		return
	}
	fullURL, err := handler.service.FindAllByUserID(userID)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(fullURL) == 0 {
		http.Error(res, "no content", http.StatusNoContent)
		return
	}
	resp, err := json.Marshal(&fullURL)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(http.StatusOK)
	_, err = res.Write(resp)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
