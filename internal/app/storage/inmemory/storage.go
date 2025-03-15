package inmemory

import "fmt"

type Storage struct {
	bucket map[string]string
}

func (s *Storage) Save(urlHash string, fullURL string) {
	s.bucket[urlHash] = fullURL
}

func (s *Storage) FindByHashURL(hashURL string) (string, error) {
	if fullURL, exists := s.bucket[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

func NewStorage() *Storage {
	return &Storage{bucket: make(map[string]string)}
}
