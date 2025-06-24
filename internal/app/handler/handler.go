package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/service"
)

// Handler — это объединяющая структура, содержащая все HTTP-обработчики приложения.
// Она предоставляет доступ к функциональности:
// - создание коротких ссылок (plain text и JSON),
// - пакетное создание,
// - поиск по хэшу и по пользователю,
// - удаление,
// - проверка состояния сервиса (Ping).
type Handler struct {
	Create
	CreateWithJSON
	Batch
	Find
	Ping
	Delete
}

// CreateHandler инициализирует и возвращает новый экземпляр Handler с заданными зависимостями.
//
// Параметры:
//   - s: указатель на сервис типа *service.Shortener, реализующий бизнес-логику.
//   - pingChecker: реализация интерфейса PingChecker для проверки состояния БД.
//   - cfg: конфигурация приложения, включающая, например, ключ аутентификации.
//
// Возвращает:
//   - *Handler: готовый к использованию объект обработчика HTTP-запросов.
func CreateHandler(s *service.Shortener, pingChecker PingChecker, cfg *config.Config) *Handler {
	return &Handler{
		Create:         Create{service: s, authKey: cfg.AuthKey},
		CreateWithJSON: CreateWithJSON{service: s, authKey: cfg.AuthKey},
		Batch:          Batch{service: s, authKey: cfg.AuthKey},
		Find:           Find{service: s, authKey: cfg.AuthKey},
		Ping:           Ping{pingChecker},
		Delete:         Delete{service: s, authKey: cfg.AuthKey},
	}
}
