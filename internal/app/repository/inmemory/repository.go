package inmemory

import (
	"fmt"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

type Repository struct {
	bucket map[string]string
	bkp    *Backup
}

func (r *Repository) Save(urlHash string, fullURL string) {
	if _, exists := r.bucket[urlHash]; !exists {
		err := r.bkp.Write(urlHash, fullURL)
		if err != nil {
			logger.Log.Error("backup writing failed", zap.Error(err))
		}
		r.bucket[urlHash] = fullURL
	}
}

func (r *Repository) FindByHash(hashURL string) (string, error) {
	if fullURL, exists := r.bucket[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

func NewInMemoryRepository(backupFilePath string) *Repository {
	bucket := make(map[string]string)
	bkp, err := NewBackup(backupFilePath)
	if err != nil {
		logger.Log.Error("create backup failed", zap.Error(err))
	} else {
		bkp.RecoverTo(bucket)
	}
	return &Repository{
		bucket: bucket,
		bkp:    bkp,
	}
}
