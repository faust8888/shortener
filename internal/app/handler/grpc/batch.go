// Package grpc реализует gRPC-сервисы для операций с URL-сокращением.
package grpc

import (
	"context"
	"fmt"
	"github.com/faust8888/shortener/internal/app/model"
	pb "github.com/faust8888/shortener/internal/app/proto/shortenerpb"
	"github.com/faust8888/shortener/internal/app/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Batch реализует gRPC сервер для пакетного создания коротких ссылок.
//
// Выполняет аутентификацию через токены, извлечённые из контекста,
// преобразует protobuf-запросы в внутренние модели,
// вызывает бизнес-логику для создания ссылок,
// и возвращает protobuf-ответы или соответствующие ошибки gRPC.
type Batch struct {
	service batchSaver
	authKey string
	pb.UnimplementedShortenerServiceServer
}

// batchSaver описывает минимальный набор методов сервиса,
// необходимых для пакетного создания коротких ссылок.
type batchSaver interface {
	// CreateWithBatch создаёт несколько коротких ссылок для указанного пользователя.
	//
	// Параметры:
	//   - batch: срез DTO с URL для сокращения и корреляционными ID.
	//   - userID: идентификатор аутентифицированного пользователя.
	//
	// Возвращает:
	//   - срез DTO с корреляционными ID и сгенерированными короткими ссылками.
	//   - ошибку, если операция завершилась неудачей.
	CreateWithBatch(batch []model.CreateShortRequestBatchItemRequest, userID string) ([]model.CreateShortRequestBatchItemResponse, error)
}

// CreateBatch обрабатывает входящий gRPC-запрос на пакетное создание коротких ссылок
// и возвращает пакетный ответ.
//
// Выполняет извлечение и валидацию токена,
// преобразует protobuf-запрос в внутренний формат,
// вызывает сервис для создания пакетных ссылок,
// формирует protobuf-ответ.
//
// В случае ошибок возвращает соответствующие gRPC коды ошибок,
// такие как Unauthenticated для проблем с аутентификацией или Internal для внутренних ошибок.
func (s *Batch) CreateBatch(ctx context.Context, req *pb.BatchCreateRequest) (*pb.BatchCreateResponse, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil || token == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized: token missing")
	}

	userID, err := security.GetUserID(token, s.authKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	batchRequests := make([]model.CreateShortRequestBatchItemRequest, len(req.Batch))
	for i, item := range req.Batch {
		batchRequests[i] = model.CreateShortRequestBatchItemRequest{
			CorrelationID: item.CorrelationId,
			OriginalURL:   item.OriginalUrl,
		}
	}

	batchResponses, err := s.service.CreateWithBatch(batchRequests, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to Create Batch: %v", err))
	}

	pbResponses := make([]*pb.CreateShortRequestBatchItemResponse, len(batchResponses))
	for i, item := range batchResponses {
		pbResponses[i] = &pb.CreateShortRequestBatchItemResponse{
			CorrelationId: item.CorrelationID,
			ShortUrl:      item.ShortURL,
		}
	}

	return &pb.BatchCreateResponse{
		Batch: pbResponses,
	}, nil
}
