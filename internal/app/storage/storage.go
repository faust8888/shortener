package storage

import "fmt"

type Storage interface {
	Save(urlHash string, fullURL string)
	FindByHashURL(shortURL string) (string, error)
}

type InMemoryStorage struct {
	mapStorage map[string]string
}

func (b *InMemoryStorage) Save(urlHash string, fullURL string) {
	b.mapStorage[urlHash] = fullURL
}

func (b *InMemoryStorage) FindByHashURL(hashURL string) (string, error) {
	if fullURL, exists := b.mapStorage[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

func NewInMemoryStorage() Storage {
	return &InMemoryStorage{mapStorage: make(map[string]string)}
}
