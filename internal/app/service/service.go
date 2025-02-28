package service

import (
	"fmt"
	"github.com/faust8888/shortener/cmd/config"
	"github.com/faust8888/shortener/internal/app/storage"
	"github.com/faust8888/shortener/internal/app/util"
)

type URLShortener interface {
	CreateShortURL(fullURL string) (string, error)
	FindFullURL(hashURL string) (string, error)
}

type URLShortenerService struct {
	storage storage.Storage
}

func (s *URLShortenerService) CreateShortURL(fullURL string) (string, error) {
	urlHash, err := s.createHashForURL(fullURL)
	if err != nil {
		return "", err
	}
	s.storage.Save(urlHash, fullURL)
	shortURL := fmt.Sprintf("%s/%s", config.Cfg.BaseShortURL, urlHash)

	return shortURL, nil
}

func (s *URLShortenerService) FindFullURL(hashURL string) (string, error) {
	return s.storage.FindByHashURL(hashURL)
}

func NewInMemoryShortenerService() *URLShortenerService {
	return &URLShortenerService{storage: storage.NewInMemoryStorage()}
}

func (s *URLShortenerService) createHashForURL(fullURL string) (string, error) {
	if util.IsInvalidURL(fullURL) {
		return "", fmt.Errorf("invalid url")
	}
	return util.CreateHash(fullURL), nil
}
