package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/faust8888/shortener/internal/app/storage"
	"net/url"
)

type URLShortener interface {
	CreateShortURL(fullURL string) (string, error)
	FindFullURL(hashURL string) (string, error)
}

type URLShortenerService struct {
	storage      storage.Storage
	baseShortURL string
}

func (s *URLShortenerService) CreateShortURL(fullURL string) (string, error) {
	urlHash, err := createHashForURL(fullURL)
	if err != nil {
		return "", err
	}
	s.storage.Save(urlHash, fullURL)
	shortURL := fmt.Sprintf("%s/%s", s.baseShortURL, urlHash)

	return shortURL, nil
}

func (s *URLShortenerService) FindFullURL(hashURL string) (string, error) {
	return s.storage.FindByHashURL(hashURL)
}

func NewURLShortener(s storage.Storage, baseShortURL string) *URLShortenerService {
	return &URLShortenerService{storage: s, baseShortURL: baseShortURL}
}

func createHashForURL(fullURL string) (string, error) {
	if isInvalidURL(fullURL) {
		return "", fmt.Errorf("invalid url")
	}
	return createHash(fullURL), nil
}

func isInvalidURL(fullURL string) bool {
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return true
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return true
	}
	return false
}

func createHash(key string) string {
	hashBytes := sha256.Sum256([]byte(key))
	hashString := base64.URLEncoding.EncodeToString(hashBytes[:])
	return hashString[:10]
}
