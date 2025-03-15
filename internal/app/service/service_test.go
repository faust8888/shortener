package service

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/storage/inmemory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"strings"
	"testing"
)

const (
	TestURL                    = "https://google.com"
	CreateShortURLErrorMessage = "Failed to create short URL"
	GetFullURLErrorMessage     = "Failed to find full URL"
	URLNotMatchErrorMessage    = "url doesn't match"
)

func TestCreatingShortURLAndFinding(t *testing.T) {
	cfg := config.Create()
	service := NewURLShortener(inmemory.NewStorage(), cfg.BaseShortURL)
	tests := []struct {
		name    string
		fullURL string
	}{
		{
			name:    "Successfully Create and Find URL",
			fullURL: "https://ya.ru",
		},
		{
			name:    "Successfully Create and Find URL (long url)",
			fullURL: "https://sven.ru/qwer/yrue/123/0393/kdjdksadasnjda/923839238/asjasjdi",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURL, err := service.CreateShortURL(test.fullURL)
			require.NoError(t, err, CreateShortURLErrorMessage)
			parsedURL, _ := url.Parse(shortURL)

			hashURL := strings.TrimPrefix(parsedURL.Path, "/")
			returnedFullURL, err := service.FindFullURL(hashURL)

			require.NoError(t, err, GetFullURLErrorMessage)
			assert.Equal(t, test.fullURL, returnedFullURL, URLNotMatchErrorMessage)
		})
	}
}

func TestCouldNotFindFullURL(t *testing.T) {
	cfg := config.Create()
	service := NewURLShortener(inmemory.NewStorage(), cfg.ServerAddress)
	_, err := service.CreateShortURL(TestURL)
	require.NoError(t, err, CreateShortURLErrorMessage)

	fullURL, err := service.FindFullURL("not_existing_hash_key")

	require.Error(t, err, "Expected error when trying to find full URL")
	require.Equal(t, "short url not found for not_existing_hash_key", err.Error())
	require.Equal(t, "", fullURL)
}
