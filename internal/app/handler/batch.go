package handler

import (
	"encoding/json"
	"fmt"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/security"
	"net/http"
)

type Batch struct {
	service batchSaver
	authKey string
}

type batchSaver interface {
	CreateWithBatch(batch []model.CreateShortRequestBatchItemRequest, userID string) ([]model.CreateShortRequestBatchItemResponse, error)
}

// CreateLinkWithBatch обрабатывает входящий POST-запрос с пакетом данных для создания коротких ссылок.
//
// Метод:
// - Проверяет или генерирует токен авторизации.
// - Извлекает идентификатор пользователя из токена.
// - Декодирует JSON-тело запроса.
// - Передаёт данные сервису для сохранения.
// - Возвращает JSON-ответ со списком созданных ссылок.
//
// Пример тела запроса:
//
//	[
//	  {"correlation_id": "id1", "original_url": "http://example.com/1"},
//	  {"correlation_id": "id2", "original_url": "http://example.com/2"}
//	]
//
// Ответ:
//
//	[
//	  {"correlation_id": "id1", "short_url": "http://your-shortener.com/abc"},
//	  {"correlation_id": "id2", "short_url": "http://your-shortener.com/def"}
//	]
//
// Возможные HTTP-статусы:
// - 201 Created — успешно созданы.
// - 400 Bad Request — невалидное тело запроса.
// - 401 Unauthorized — отсутствующий или недействительный токен.
// - 500 Internal Server Error — внутренняя ошибка сервера.
func (handler *Batch) CreateLinkWithBatch(res http.ResponseWriter, req *http.Request) {
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
