package service

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/repository/inmemory"
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
	service := CreateShortener(inmemory.NewInMemoryRepository(cfg), cfg.BaseShortURL)
	tests := []struct {
		name    string
		fullURL string
		userID  string
	}{
		{
			name:    "Successfully CreateLink and FindLinkByHash URL",
			fullURL: "https://ya.ru",
			userID:  "123456",
		},
		{
			name:    "Successfully CreateLink and FindLinkByHash URL (long url)",
			fullURL: "https://sven.ru/qwer/yrue/123/0393/kdjdksadasnjda/923839238/asjasjdi",
			userID:  "123456",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURL, err := service.Create(test.fullURL, test.userID)
			require.NoError(t, err, CreateShortURLErrorMessage)
			parsedURL, _ := url.Parse(shortURL)

			hashURL := strings.TrimPrefix(parsedURL.Path, "/")
			returnedFullURL, err := service.FindByHash(hashURL)

			require.NoError(t, err, GetFullURLErrorMessage)
			assert.Equal(t, test.fullURL, returnedFullURL, URLNotMatchErrorMessage)
		})
	}
}

func TestCouldNotFindFullURL(t *testing.T) {
	cfg := config.Create()
	shortener := CreateShortener(inmemory.NewInMemoryRepository(cfg), cfg.ServerAddress)
	_, err := shortener.Create(TestURL, "123456")
	require.NoError(t, err, CreateShortURLErrorMessage)

	fullURL, err := shortener.FindByHash("not_existing_hash_key")

	require.Error(t, err, "Expected error when trying to find full URL")
	require.Equal(t, "find by hash: short url not found for not_existing_hash_key", err.Error())
	require.Equal(t, "", fullURL)
}

func BenchmarkCreatingShortURLAndFinding(b *testing.B) {
	b.StopTimer()
	cfg := config.Create()
	service := CreateShortener(inmemory.NewInMemoryRepository(cfg), cfg.BaseShortURL)
	tests := []struct {
		name    string
		fullURL string
		userID  string
	}{
		{
			name:    "Successfully CreateLink and FindLinkByHash URL",
			fullURL: "https://ya.ru",
			userID:  "123456",
		},
		{
			name:    "Successfully CreateLink and FindLinkByHash URL (long url)",
			fullURL: "https://sven.ru/qwer/yrue/123/0393/kdjdksadasnjda/923839238/asjasjdi",
			userID:  "123456",
		},
	}
	b.StartTimer()
	for _, test := range tests {
		b.Run(test.name, func(t *testing.B) {
			shortURL, err := service.Create(test.fullURL, test.userID)
			require.NoError(t, err, CreateShortURLErrorMessage)
			parsedURL, _ := url.Parse(shortURL)

			hashURL := strings.TrimPrefix(parsedURL.Path, "/")
			returnedFullURL, err := service.FindByHash(hashURL)

			require.NoError(t, err, GetFullURLErrorMessage)
			assert.Equal(t, test.fullURL, returnedFullURL, URLNotMatchErrorMessage)
		})
	}
}
