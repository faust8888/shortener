package inmemory

import (
	"fmt"
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

type Repository struct {
	urlBucket    map[string]string
	userBucket   map[string]map[string]struct{}
	bkp          *Backup
	baseShortURL string
}

func (r *Repository) Save(urlHash string, fullURL string, userID string) error {
	if _, exists := r.urlBucket[urlHash]; !exists {
		r.urlBucket[urlHash] = fullURL
	}
	if _, exists := r.userBucket[userID]; !exists {
		r.userBucket[userID] = make(map[string]struct{})
	}
	err := r.bkp.Write(urlHash, fullURL, userID)
	if err != nil {
		logger.Log.Error("backup writing failed", zap.Error(err))
	}
	r.userBucket[userID][urlHash] = struct{}{}
	return nil
}

func (r *Repository) FindByHash(hashURL string) (string, error) {
	if fullURL, exists := r.urlBucket[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

func (r *Repository) FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error) {
	shortURLs := r.userBucket[userID]
	result := make([]model.FindURLByUserIDResponse, 0)
	for shortURL := range shortURLs {
		originalURL := r.urlBucket[shortURL]
		result = append(result, model.FindURLByUserIDResponse{
			OriginalURL: originalURL,
			ShortURL:    fmt.Sprintf("%s/%s", r.baseShortURL, shortURL),
		})
	}
	return result, nil
}

func (r *Repository) SaveAll(batch map[string]model.CreateShortDTO, userID string) error {
	for _, batchItem := range batch {
		err := r.Save(batchItem.HashURL, batchItem.OriginalURL, userID)
		if err != nil {
			return fmt.Errorf("inmemory.repository.saveAll: %w", err)
		}
	}
	return nil
}

func (r *Repository) Ping() (bool, error) {
	return true, nil
}

func (r *Repository) DeleteAll(shortURLs []string, userID string) error {
	return nil
}

func NewInMemoryRepository(cfg *config.Config) *Repository {
	urlBucket := make(map[string]string)
	userBucket := make(map[string]map[string]struct{})
	bkp, err := NewBackup(cfg.StorageFilePath)
	if err != nil {
		logger.Log.Error("create backup failed", zap.Error(err))
	} else {
		bkp.RecoverTo(urlBucket, userBucket)
	}
	return &Repository{
		urlBucket:    urlBucket,
		userBucket:   userBucket,
		bkp:          bkp,
		baseShortURL: cfg.BaseShortURL,
	}
}
