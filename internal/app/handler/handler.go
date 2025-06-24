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
// - проверка состояния сервиса (ping).
type Handler struct {
	create
	createWithJSON
	batch
	find
	ping
	delete
}

// Create инициализирует и возвращает новый экземпляр Handler с заданными зависимостями.
//
// Параметры:
//   - s: указатель на сервис типа *service.Shortener, реализующий бизнес-логику.
//   - pingChecker: реализация интерфейса PingChecker для проверки состояния БД.
//   - cfg: конфигурация приложения, включающая, например, ключ аутентификации.
//
// Возвращает:
//   - *Handler: готовый к использованию объект обработчика HTTP-запросов.
func Create(s *service.Shortener, pingChecker PingChecker, cfg *config.Config) *Handler {
	return &Handler{
		create:         create{service: s, authKey: cfg.AuthKey},
		createWithJSON: createWithJSON{service: s, authKey: cfg.AuthKey},
		batch:          batch{service: s, authKey: cfg.AuthKey},
		find:           find{service: s, authKey: cfg.AuthKey},
		ping:           ping{pingChecker},
		delete:         delete{service: s, authKey: cfg.AuthKey},
	}
}
