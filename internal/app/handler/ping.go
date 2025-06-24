package handler

import (
	"net/http"
)

// Ping — это HTTP-обработчик для проверки состояния сервиса.
// Используется для мониторинга доступности приложения и его зависимостей (например, БД).
type Ping struct {
	service PingChecker
}

// PingChecker — интерфейс, определяющий метод для проверки работоспособности сервиса.
// Реализация должна возвращать true и nil, если сервис доступен.
type PingChecker interface {
	Ping() (bool, error)
}

// PingDatabase обрабатывает GET-запрос на эндпоинт /PingDatabase и проверяет доступность сервиса.
//
// Метод:
// - Вызывает Ping() у сервиса.
// - Если ошибка отсутствует — возвращает 200 OK.
// - Если есть ошибка — возвращает 500 Internal Server Error и текст ошибки.
//
// Путь: /PingDatabase
//
// Возможные HTTP-статусы:
// - 200 OK — сервис доступен.
// - 500 Internal Server Error — внутренняя ошибка или недоступная зависимость.
func (handler *Ping) PingDatabase(res http.ResponseWriter, req *http.Request) {
	_, err := handler.service.Ping()
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
	res.WriteHeader(http.StatusOK)
}
