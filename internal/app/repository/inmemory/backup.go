package inmemory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/faust8888/shortener/internal/middleware/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"os"
)

type Backup struct {
	file    *os.File
	writer  *bufio.Writer
	scanner *bufio.Scanner
}

func (p *Backup) Write(urlHash string, fullURL string) error {
	backupEvent := &createShortBackupEvent{
		UUID:        uuid.New(),
		ShortURL:    urlHash,
		OriginalURL: fullURL,
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

func (p *Backup) RecoverTo(bucket map[string]string) {
	for p.scanner.Scan() {
		event := createShortBackupEvent{}
		err := json.Unmarshal(p.scanner.Bytes(), &event)
		if err != nil {
			logger.Log.Error("failed to unmarshal a backup event", zap.Error(err))
		} else {
			logger.Log.Info("recovering backup event", zap.Any("event", event))
		}
		bucket[event.ShortURL] = event.OriginalURL
	}
}

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

type createShortBackupEvent struct {
	UUID        uuid.UUID `json:"uuid" validate:"required,uuid"`
	ShortURL    string    `json:"short_url" validate:"required,short_url"`
	OriginalURL string    `json:"original_url" validate:"required,original_url"`
}

func (e createShortBackupEvent) String() string {
	return fmt.Sprintf(`{"UUID": "%s", "ShortURL": "%s", "OriginalURL": "%s"}`,
		e.UUID, e.ShortURL, e.OriginalURL)
}
