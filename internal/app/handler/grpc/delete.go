// Package grpc реализует gRPC-обработчики для операций с URL-сокращением.
package grpc

import (
	"context"
	pb "github.com/faust8888/shortener/internal/app/proto/shortenerpb"
	"github.com/faust8888/shortener/internal/app/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Delete реализует gRPC сервер для асинхронного удаления коротких ссылок.
type Delete struct {
	service deleter
	authKey string
}

// deleter определяет интерфейс сервиса, поддерживающего асинхронное удаление ссылок.
type deleter interface {
	// DeleteAsync запускает асинхронное удаление ссылок по ID для указанного пользователя.
	DeleteAsync(ids []string, userID string) error
}

// DeleteLinks обрабатывает gRPC запрос на удаление нескольких коротких ссылок.
// Метод извлекает токен из контекста, проверяет аутентификацию,
// вызывает сервис асинхронного удаления и возвращает пустой ответ при успешном приёме.
func (d *Delete) DeleteLinks(ctx context.Context, req *pb.DeleteLinksRequest) (*pb.DeleteLinksResponse, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil || token == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized: token missing")
	}
	userID, err := security.GetUserID(token, d.authKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	if err = d.service.DeleteAsync(req.GetIds(), userID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete links: %v", err)
	}

	// Возвращаем пустой ответ как подтверждение принятия запроса на удаление.
	return &pb.DeleteLinksResponse{}, nil
}
