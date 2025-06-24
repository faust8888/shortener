package handler

import (
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/security"
	"io"
	"net/http"
)

type Delete struct {
	service deleter
	authKey string
}

type deleter interface {
	DeleteAsync(ids []string, userID string) error
}

// DeleteLink обрабатывает POST-запрос на удаление нескольких коротких ссылок.
//
// Метод:
// - Проверяет наличие токена авторизации.
// - Извлекает идентификатор пользователя из токена.
// - Читает и декодирует JSON-тело запроса, ожидая массив строк (ID ссылок).
// - Передаёт данные сервису для асинхронного удаления.
//
// Пример тела запроса:
//
//	["abc123", "def456"]
//
// Возможные HTTP-статусы:
// - 202 Accepted — запрос принят на обработку (асинхронное удаление).
// - 400 Bad Request — невалидное тело запроса или ошибка парсинга.
// - 401 Unauthorized — отсутствующий или недействительный токен.
// - 500 Internal Server Error — внутренняя ошибка сервера.
func (handler *Delete) DeleteLink(res http.ResponseWriter, req *http.Request) {
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
