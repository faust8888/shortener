// Package grpc содержит реализацию gRPC-сервера для получения статистики по сокращённым URL.
package grpc

import (
	"context"
	"github.com/faust8888/shortener/internal/app/model"
	pb "github.com/faust8888/shortener/internal/app/proto/shortenerpb"
	"github.com/faust8888/shortener/internal/app/security"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StatServer реализует gRPC сервис для получения статистики.
// Выполняет проверку IP клиента на принадлежность доверенной подсети.
type StatServer struct {
	pb.UnimplementedShortenerServiceServer
	service       Collector
	trustedSubnet string
}

// Collector описывает интерфейс сервиса, возвращающего статистику.
type Collector interface {
	// Collect возвращает статистические данные по коротким URL.
	Collect() (*model.Statistic, error)
}

// GetStatistics проверяет IP клиента из gRPC метаданных на вхождение в доверенную подсеть,
// и если IP разрешён, возвращает статистику.
// В случае ошибки возвращает соответствующий gRPC статус с кодом.
func (s *StatServer) GetStatistics(ctx context.Context, req *pb.EmptyRequest) (*pb.Statistic, error) {
	_, err := security.IsAllowedTrustedIPFromContext(ctx, s.trustedSubnet)
	if err != nil {
		return nil, err
	}

	stat, err := s.service.Collect()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to collect stats: %v", err)
	}

	return &pb.Statistic{
		Urls:  int32(stat.Urls),
		Users: int32(stat.Users),
	}, nil
}
