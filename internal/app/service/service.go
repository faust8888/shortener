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

// Shortener — это основной сервис приложения, реализующий бизнес-логику для работы с короткими ссылками.
// Содержит зависимости от репозитория и базового URL.
type Shortener struct {
	repository   repository.Repository // Интерфейс хранилища для операций над данными
	baseShortURL string                // Базовый URL для формирования полного адреса короткой ссылки
}

// Create создаёт новую короткую ссылку на основе оригинального URL.
//
// Параметры:
//   - fullURL: оригинальный URL.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - string: готовая короткая ссылка.
//   - error: nil, если успешно, иначе — ошибку.
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

// FindByHash находит оригинальный URL по его хэш-ключу.
//
// Параметры:
//   - hashURL: хэш-ключ короткой ссылки.
//
// Возвращает:
//   - string: оригинальный URL.
//   - error: nil, если найдено, иначе — ошибку.
func (s *Shortener) FindByHash(hashURL string) (string, error) {
	foundURL, err := s.repository.FindByHash(hashURL)
	if err != nil {
		logger.Log.Error("couldn't find short URL", zap.Error(err))
		return "", fmt.Errorf("find by hash: %w", err)
	}
	logger.Log.Info("found short URL", zap.String("hashURL", hashURL))
	return foundURL, nil
}

// FindAllByUserID возвращает все короткие ссылки, принадлежащие пользователю.
//
// Параметр:
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []model.FindURLByUserIDResponse: список ссылок пользователя.
//   - error: nil, если успешно, иначе — ошибку.
func (s *Shortener) FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error) {
	r, err := s.repository.FindAllByUserID(userID)
	if err != nil {
		return []model.FindURLByUserIDResponse{}, fmt.Errorf("find all by user id (%s): %w", userID, err)
	}
	return r, nil
}

// CreateWithBatch создаёт несколько коротких ссылок за один раз (пакетная операция).
//
// Параметры:
//   - batch: массив элементов запроса с correlation_id и original_url.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []model.CreateShortRequestBatchItemResponse: массив результатов с correlation_id и short_url.
//   - error: nil, если успешно, иначе — ошибку.
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

// DeleteAsync удаляет несколько коротких ссылок асинхронно.
//
// Принимает список идентификаторов (хэшей) и идентификатор пользователя.
// Операция выполняется в отдельной горутине.
//
// Параметры:
//   - ids: список идентификаторов (хэшей) ссылок для удаления.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: всегда nil, так как ошибка внутри горутины не возвращается.
func (s *Shortener) DeleteAsync(ids []string, userID string) error {
	go func(ids []string) {
		err := s.repository.DeleteAll(ids, userID)
		if err != nil {
			logger.Log.Info("couldn't delete URLs")
		}
	}(ids)
	return nil
}

// CreateShortener инициализирует и возвращает новый экземпляр сервиса Shortener.
//
// Параметры:
//   - s: реализация интерфейса repository.Repository.
//   - baseShortURL: базовый URL для формирования полных адресов коротких ссылок.
//
// Возвращает:
//   - *Shortener: готовый к использованию объект сервиса.
func CreateShortener(s repository.Repository, baseShortURL string) *Shortener {
	return &Shortener{repository: s, baseShortURL: baseShortURL}
}

// createBatchMap преобразует пакет входящих данных в карту DTO для сохранения.
//
// Используется внутренне в методе CreateWithBatch.
//
// Параметры:
//   - batch: массив элементов запроса.
//
// Возвращает:
//   - map[string]model.CreateShortDTO: карта correlation_id → DTO.
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
