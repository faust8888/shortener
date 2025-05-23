package service

import (
	"errors"
	"fmt"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/repository"
	"github.com/faust8888/shortener/internal/app/repository/postgres"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

type Shortener struct {
	repository   repository.Repository
	baseShortURL string
}

func (s *Shortener) Create(fullURL, userID string) (string, error) {
	urlHash, err := security.CreateHashForURL(fullURL)
	if err != nil {
		return "", fmt.Errorf("hash for url: %w", err)
	}
	err = s.repository.Save(urlHash, fullURL, userID)
	if err != nil && !errors.Is(err, postgres.ErrUniqueIndexConstraint) {
		return "", fmt.Errorf("saving data: %w", err)
	}
	shortURL := fmt.Sprintf("%s/%s", s.baseShortURL, urlHash)
	logger.Log.Info("created short URL", zap.String("shortUrl", shortURL), zap.String("fullUrl", fullURL))
	return shortURL, err
}

func (s *Shortener) FindByHash(hashURL string) (string, error) {
	foundURL, err := s.repository.FindByHash(hashURL)
	if err != nil {
		logger.Log.Error("couldn't find short URL", zap.Error(err))
		return "", fmt.Errorf("find by hash: %w", err)
	}
	logger.Log.Info("found short URL", zap.String("hashURL", hashURL))
	return foundURL, nil
}

func (s *Shortener) FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error) {
	r, err := s.repository.FindAllByUserID(userID)
	if err != nil {
		return []model.FindURLByUserIDResponse{}, fmt.Errorf("find all by user id (%s): %w", userID, err)
	}
	return r, nil
}

func (s *Shortener) CreateWithBatch(batch []model.CreateShortRequestBatchItemRequest, userID string) ([]model.CreateShortRequestBatchItemResponse, error) {
	logger.Log.Info("creating short URLs with batch", zap.Int("size", len(batch)))
	batchMap := s.createBatchMap(batch)
	err := s.repository.SaveAll(batchMap, userID)
	if err != nil {
		return nil, fmt.Errorf("service.createWithBatch: %w", err)
	}
	var result = make([]model.CreateShortRequestBatchItemResponse, 0)
	for correlationID, value := range batchMap {
		logger.Log.Info("created",
			zap.String("correlationID", correlationID),
			zap.String("shortURL", value.ShortURL),
			zap.String("originalURL", value.OriginalURL))
		result = append(result,
			model.CreateShortRequestBatchItemResponse{
				CorrelationID: correlationID,
				ShortURL:      value.ShortURL,
			})
	}
	return result, nil
}

func (s *Shortener) DeleteAsync(ids []string, userID string) error {
	go func(ids []string) {
		err := s.repository.DeleteAll(ids, userID)
		if err != nil {
			logger.Log.Info("couldn't delete URLs")
		}
	}(ids)
	return nil
}

func CreateShortener(s repository.Repository, baseShortURL string) *Shortener {
	return &Shortener{repository: s, baseShortURL: baseShortURL}
}

func (s *Shortener) createBatchMap(batch []model.CreateShortRequestBatchItemRequest) map[string]model.CreateShortDTO {
	var createShortMap = make(map[string]model.CreateShortDTO)
	for _, batchItem := range batch {
		var hashURL = security.CreateHash(batchItem.OriginalURL)
		createShortMap[batchItem.CorrelationID] = model.CreateShortDTO{
			OriginalURL: batchItem.OriginalURL,
			ShortURL:    fmt.Sprintf("%s/%s", s.baseShortURL, hashURL),
			HashURL:     hashURL,
		}
	}
	return createShortMap
}
