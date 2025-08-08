// Package grpc реализует gRPC сервер для операций поиска и получения ссылок.
package grpc

import (
	"context"
	"errors"
	"github.com/faust8888/shortener/internal/app/model"
	pb "github.com/faust8888/shortener/internal/app/proto/shortenerpb"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/app/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Find реализует gRPC-сервер для поиска коротких ссылок и получения всех ссылок пользователя.
type Find struct {
	service finder
	authKey string
}

// finder описывает интерфейс сервиса для поиска URL по хэшу и получения всех URL пользователя.
type finder interface {
	FindByHash(hashURL string) (string, error)
	FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error)
}

// FindByHash обрабатывает gRPC-запрос на поиск полного URL по короткому хэшу.
//
// Возвращает:
// - URL и статус отсутствия удаления, если запись найдена и не удалена;
// - статус "удалено", если запись помечена удалённой;
// - ошибку NOT_FOUND, если запись не найдена.
func (s *Find) FindByHash(ctx context.Context, req *pb.FindByHashRequest) (*pb.FindByHashResponse, error) {
	fullURL, err := s.service.FindByHash(req.GetHash())
	if errors.Is(err, postgres.ErrRecordWasMarkedAsDeleted) {
		return &pb.FindByHashResponse{
			Status: pb.DeletedStatus_MARKED_DELETED,
		}, nil
	}
	if err != nil {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &pb.FindByHashResponse{
		FullUrl: fullURL,
		Status:  pb.DeletedStatus_NONE,
	}, nil
}

// FindAllByUserID обрабатывает gRPC-запрос на получение всех коротких ссылок пользователя.
//
// Метод извлекает токен из контекста, проверяет его и получает userID,
// запрашивает все ссылки у сервиса,
// и возвращает их или соответствующую ошибку при отсутствии или внутренних ошибках.
func (s *Find) FindAllByUserID(ctx context.Context, _ *pb.FindAllByUserIDRequest) (*pb.FindAllByUserIDResponse, error) {
	token, err := security.GetTokenFromContext(ctx)
	if err != nil || token == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	userID, err := security.GetUserID(token, s.authKey)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	urls, err := s.service.FindAllByUserID(userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "no content")
	}

	pbURLs := make([]*pb.FindURLByUserIDResponse, 0, len(urls))
	for _, u := range urls {
		pbURLs = append(pbURLs, &pb.FindURLByUserIDResponse{
			ShortUrl:    u.ShortURL,
			OriginalUrl: u.OriginalURL,
		})
	}

	return &pb.FindAllByUserIDResponse{
		Urls: pbURLs,
	}, nil
}
