package service

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/faust8888/shortener/internal/app/repository"
	"net/url"
)

type Shortener struct {
	repository   repository.Repository
	baseShortURL string
}

func (s *Shortener) Create(fullURL string) (string, error) {
	urlHash, err := createHashForURL(fullURL)
	if err != nil {
		return "", err
	}
	s.repository.Save(urlHash, fullURL)
	shortURL := fmt.Sprintf("%s/%s", s.baseShortURL, urlHash)

	return shortURL, nil
}

func (s *Shortener) Find(hashURL string) (string, error) {
	return s.repository.FindByHash(hashURL)
}

func CreateShortener(s repository.Repository, baseShortURL string) *Shortener {
	return &Shortener{repository: s, baseShortURL: baseShortURL}
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
