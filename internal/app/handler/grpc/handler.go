// Package grpc содержит реализацию gRPC сервера для сервиса сокращения ссылок.
package grpc

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/repository"
	"github.com/faust8888/shortener/internal/app/service"
)

// ShortenerServer агрегирует все части gRPC сервиса для коротких ссылок:
// создание, поиск, пакетные операции и удаление.
type ShortenerServer struct {
	Create
	Find
	Batch
	Delete
}

// CreateServer создает и инициализирует новый экземпляр ShortenerServer.
//
// Аргументы:
//   - s: экземпляр бизнес-логики сервиса сокращения ссылок.
//   - repo: репозиторий для работы с данными (обычно передается в сервис).
//   - cfg: конфигурация приложения с параметрами, включая ключ авторизации.
//
// Возвращает подготовленный к работе gRPC-сервер, объединяющий все необходимые обработчики.
func CreateServer(s *service.Shortener, repo repository.Repository, cfg *config.Config) *ShortenerServer {
	return &ShortenerServer{
		Create: Create{service: s, authKey: cfg.AuthKey},
		Find:   Find{service: s, authKey: cfg.AuthKey},
		Batch:  Batch{service: s, authKey: cfg.AuthKey},
		Delete: Delete{service: s, authKey: cfg.AuthKey},
	}
}
