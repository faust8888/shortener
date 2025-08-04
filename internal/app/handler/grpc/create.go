// Package grpc реализует обработчики gRPC-сервисов для операций с сокращением URL.
package grpc

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/faust8888/shortener/internal/app/proto/shortenerpb"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

// Create реализует gRPC-обработчик для создания коротких ссылок.
//
// Обрабатывает извлечение и проверку токена авторизации из метаданных контекста,
// взаимодействует с сервисом для генерации короткой ссылки,
// возвращает соответствующие ошибки при неавторизованном доступе, нарушениях уникальности и других ошибках.
type Create struct {
	service creator
	authKey string
}

// creator описывает интерфейс сервиса, реализующего генерацию коротких ссылок.
type creator interface {
	// Create создаёт короткую ссылку для заданного полного URL и идентификатора пользователя.
	Create(fullURL string, userID string) (string, error)
}

// CreateShort обрабатывает gRPC-запрос на создание короткой ссылки.
//
// Метод извлекает токен из метаданных gRPC контекста,
// проверяет его, получает userID,
// вызывает сервис для создания короткой ссылки,
// дифференцирует ошибки уникальности и другие ошибки,
// и логирует ошибки при необходимости.
func (s *Create) CreateShort(ctx context.Context, req *pb.CreateShortRequest) (*pb.CreateShortResponse, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil || token == "" {
		// В gRPC обычно не создают токен здесь — аутентификация проводится отдельно,
		// но в данном коде предусмотрено создание токена при отсутствии.
		token, err = security.BuildToken(s.authKey)
		if err != nil {
			return nil, fmt.Errorf("build token: %w", err)
		}
	}

	userID, err := security.GetUserID(token, s.authKey)
	if err != nil {
		return nil, fmt.Errorf("unauthorized: %w", err)
	}

	fullURL := req.GetUrl()
	shortURL, err := s.service.Create(fullURL, userID)
	isUniqueConstraintViolation := errors.Is(err, postgres.ErrUniqueIndexConstraint)
	if err != nil && !isUniqueConstraintViolation {
		logger.Log.Error("Failed to Create short URL", zap.String("body", fullURL), zap.Error(err))
		return nil, fmt.Errorf("bad request: %w", err)
	}

	if isUniqueConstraintViolation {
		return nil, fmt.Errorf("already exists: %w", err)
	}

	return &pb.CreateShortResponse{Result: shortURL}, nil
}
