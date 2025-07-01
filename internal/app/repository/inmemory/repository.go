package inmemory

import (
	"fmt"
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
)

// Repository — это реализация интерфейса repository.Repository на основе map.
// Поддерживает:
// - хранение пар shortURL → fullURL,
// - хранение ссылок по пользователю,
// - бэкап данных в файл.
type Repository struct {
	urlBucket    map[string]string              // Карта коротких URL → оригинальные URL
	userBucket   map[string]map[string]struct{} // Карта пользовательских ссылок
	bkp          *Backup                        // Утилита для сохранения данных
	baseShortURL string                         // Базовый URL для формирования полного адреса
}

// Save сохраняет одну пару (hashURL -> fullURL) для указанного пользователя.
//
// Параметры:
//   - urlHash: хэш-ключ для короткой ссылки.
//   - fullURL: оригинальный URL.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: nil, если успешно, иначе — ошибку.
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

// FindByHash находит оригинальный URL по его хэш-ключу.
//
// Параметр:
//   - hashURL: хэш-ключ короткой ссылки.
//
// Возвращает:
//   - string: оригинальный URL.
//   - error: nil, если найдено, иначе — ошибку.
func (r *Repository) FindByHash(hashURL string) (string, error) {
	if fullURL, exists := r.urlBucket[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

// FindAllByUserID возвращает все короткие ссылки, принадлежащие пользователю.
//
// Параметр:
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []model.FindURLByUserIDResponse: список ссылок пользователя.
//   - error: nil, если успешно, иначе — ошибку.
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

// SaveAll сохраняет несколько ссылок за один раз (пакетная операция).
//
// Параметры:
//   - batch: карта хэшей и DTO с данными о ссылках.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: nil, если успешно, иначе — ошибку.
func (r *Repository) SaveAll(batch map[string]model.CreateShortDTO, userID string) error {
	for _, batchItem := range batch {
		err := r.Save(batchItem.HashURL, batchItem.OriginalURL, userID)
		if err != nil {
			return fmt.Errorf("inmemory.repository.saveAll: %w", err)
		}
	}
	return nil
}

// Ping проверяет доступность хранилища.
//
// Всегда возвращает true и nil, так как InMemory-реализация всегда доступна.
//
// Возвращает:
//   - bool: true, если хранилище доступно.
//   - error: nil, если всё в порядке.
func (r *Repository) Ping() (bool, error) {
	return true, nil
}

// DeleteAll асинхронно удаляет несколько коротких ссылок пользователя.
//
// Для InMemory-реализации пока не используется.
// Реализация удаления может быть добавлена при необходимости.
//
// Параметры:
//   - shortURLs: список идентификаторов (хэшей) ссылок для удаления.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: nil, так как удаление пока не реализовано.
func (r *Repository) DeleteAll(shortURLs []string, userID string) error {
	return nil
}

// NewInMemoryRepository создаёт новый экземпляр InMemory-репозитория.
// Если указан путь к файлу бэкапа — восстанавливает данные из него.
//
// Параметр:
//   - cfg: конфигурация приложения.
//
// Возвращает:
//   - *Repository: готовый к использованию репозиторий.
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
