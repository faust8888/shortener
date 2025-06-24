package inmemory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"go.uber.org/zap"
	"os"
)

// Backup — это утилита для сохранения и восстановления коротких ссылок в файл.
// Используется как простое решение для персистентности данных.
type Backup struct {
	file    *os.File       // Открытый файл для хранения бэкапа
	writer  *bufio.Writer  // Писатель для записи данных в файл
	scanner *bufio.Scanner // Сканер для чтения данных из файла
}

// Write записывает событие создания короткой ссылки в файл бэкапа.
//
// Параметры:
//   - urlHash: хэш-ключ (короткий URL)
//   - fullURL: оригинальный URL
//   - userID: идентификатор пользователя
//
// Возвращает:
//   - error: nil, если успешно, иначе — ошибку.
func (p *Backup) Write(urlHash, fullURL, userID string) error {
	backupEvent := &CreateShortBackupEvent{
		ShortURL:    urlHash,
		OriginalURL: fullURL,
		UserID:      userID,
	}
	data, err := json.Marshal(&backupEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal a backup event: %w", err)
	}
	if _, err = p.writer.Write(data); err != nil {
		return fmt.Errorf("failed to write data for a backup: %w", err)
	}
	if err = p.writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("failed to write new line symbol for a backup: %w", err)
	}
	return p.writer.Flush()
}

// RecoverTo восстанавливает данные из файла бэкапа в указанное хранилище.
//
// Параметры:
//   - bucket: карта для хранения пар shortURL → originalURL
//   - userBucket: карта пользовательских ссылок
func (p *Backup) RecoverTo(bucket map[string]string, userBucket map[string]map[string]struct{}) {
	for p.scanner.Scan() {
		event := CreateShortBackupEvent{}
		err := json.Unmarshal(p.scanner.Bytes(), &event)
		if err != nil {
			logger.Log.Error("failed to unmarshal a backup event", zap.Error(err))
		} else {
			logger.Log.Info("recovering backup event", zap.Any("event", event))
		}
		bucket[event.ShortURL] = event.OriginalURL
		if _, exists := userBucket[event.UserID]; exists {
			userBucket[event.UserID][event.ShortURL] = struct{}{}
		} else {
			userBucket[event.UserID] = make(map[string]struct{})
			userBucket[event.UserID][event.ShortURL] = struct{}{}
		}
	}
}

// NewBackup создаёт новый экземпляр Backup на основе указанного файла.
//
// Параметр:
//   - filename: имя файла для хранения бэкапа.
//
// Возвращает:
//   - *Backup: готовый к использованию объект бэкапа.
//   - error: nil, если успешно, иначе — ошибку.
func NewBackup(filename string) (*Backup, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("can't open file: %w", err)
	}

	return &Backup{
		file:    file,
		writer:  bufio.NewWriter(file),
		scanner: bufio.NewScanner(file),
	}, nil
}

// CreateShortBackupEvent — модель события, представляющего создание короткой ссылки.
// Хранит информацию о коротком URL, оригинальном URL и пользователе.
type CreateShortBackupEvent struct {
	ShortURL    string `json:"short_url" validate:"required,short_url"`
	OriginalURL string `json:"original_url" validate:"required,original_url"`
	UserID      string `json:"user_id" validate:"required,user_id"`
}

// String возвращает строковое представление события.
func (e CreateShortBackupEvent) String() string {
	return fmt.Sprintf(`{"ShortURL": "%s", "OriginalURL": "%s", "UserID": "%s"}`,
		e.ShortURL, e.OriginalURL, e.UserID)
}
