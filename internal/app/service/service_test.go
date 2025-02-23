package service

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"strings"
	"testing"
)

func TestCreatingShortURLAndFinding(t *testing.T) {
	service := NewInMemoryShortenerService()
	tests := []struct {
		name    string
		fullURL string
	}{
		{
			name:    "Successfully Create and Find URL (https://yandex.ru)",
			fullURL: "https://yandex.ru",
		},
		{
			name:    "Successfully Create and Find URL (long url)",
			fullURL: "http://sven.ru/qwer/yrue/123/0393/kdjdksadasnjda/923839238/asjasjdi",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURL, err := service.CreateShortURL(test.fullURL)
			require.NoError(t, err, "Failed to create short URL")
			parsedURL, _ := url.Parse(shortURL)

			returnedFullURL, err := service.FindFullURL(strings.TrimPrefix(parsedURL.Path, "/"))
			require.NoError(t, err, "Failed to find full URL")
			assert.Equal(t, test.fullURL, returnedFullURL, "Full URL does not match")
		})
	}
}

func TestCouldNotFindFullURL(t *testing.T) {
	service := NewInMemoryShortenerService()
	service.CreateShortURL("www.google.com")

	fullURL, err := service.FindFullURL("not_existing_hash_key")
	require.Error(t, err, "Expected error when trying to find full URL")
	require.Equal(t, "short url not found for not_existing_hash_key", err.Error())
	require.Equal(t, "", fullURL)
}
