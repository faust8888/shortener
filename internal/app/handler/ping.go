package handler

import (
	"net/http"
)

type ping struct {
	service PingChecker
}

// PingChecker — интерфейс, определяющий метод для проверки работоспособности сервиса.
// Реализация должна возвращать true и nil, если сервис доступен.
type PingChecker interface {
	Ping() (bool, error)
}

// Ping обрабатывает GET-запрос на эндпоинт /ping и проверяет доступность сервиса.
//
// Метод:
// - Вызывает Ping() у сервиса.
// - Если ошибка отсутствует — возвращает 200 OK.
// - Если есть ошибка — возвращает 500 Internal Server Error и текст ошибки.
//
// Путь: /ping
//
// Возможные HTTP-статусы:
// - 200 OK — сервис доступен.
// - 500 Internal Server Error — внутренняя ошибка или недоступная зависимость.
func (handler *ping) Ping(res http.ResponseWriter, req *http.Request) {
	_, err := handler.service.Ping()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}
