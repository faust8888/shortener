package storage

import "fmt"

type Storage interface {
	Save(urlHash string, fullURL string)
	FindByHashURL(hashURL string) (string, error)
}

type InMemoryStorage struct {
	mapStorage map[string]string
}

func (s *InMemoryStorage) Save(urlHash string, fullURL string) {
	s.mapStorage[urlHash] = fullURL
}

func (s *InMemoryStorage) FindByHashURL(hashURL string) (string, error) {
	if fullURL, exists := s.mapStorage[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

func NewInMemoryStorage() Storage {
	return &InMemoryStorage{mapStorage: make(map[string]string)}
}
