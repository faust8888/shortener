package http

import (
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestFindByHash(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name          string
		targetFullURL string
		want          want
	}{
		{
			name:          "Successfully get the full body by the short body",
			targetFullURL: "https://yandex.ru",
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURLResponse, _ := createShortURLRequest(server.URL, test.targetFullURL).Send()

			getFullURLRequest := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy()).R()
			getFullURLRequest.Method = http.MethodGet
			getFullURLRequest.URL = server.URL + extractHashKeyURLFrom(string(shortURLResponse.Body()))

			getFullURLResponse, _ := getFullURLRequest.Send()

			var findByUserIDResponse []model.FindURLByUserIDResponse
			_ = json.Unmarshal(getFullURLResponse.Body(), &findByUserIDResponse)

			assert.NotEmpty(t, getTokenFromResponse(shortURLResponse))
			assert.Equal(t, test.want.code, getFullURLResponse.StatusCode())

			assert.Equal(t, test.targetFullURL, getFullURLResponse.Header().Get(LocationHeader))
		})
	}
}

func TestFindByUserID(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name          string
		targetFullURL string
		want          want
	}{
		{
			name:          "Successfully get the full body by the short body",
			targetFullURL: "https://yandex.ru",
			want: want{
				code: http.StatusOK,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURLResponse, _ := createShortURLRequest(server.URL, test.targetFullURL).Send()

			getFullURLRequest := resty.New().R()
			getFullURLRequest.Method = http.MethodGet
			getFullURLRequest.URL = server.URL + "/api/user/urls"
			getFullURLRequest.SetCookies(shortURLResponse.Cookies())

			getFullURLResponse, _ := getFullURLRequest.Send()
			var findByUserIDResponse []model.FindURLByUserIDResponse
			_ = json.Unmarshal(getFullURLResponse.Body(), &findByUserIDResponse)

			assert.NotEmpty(t, getTokenFromResponse(shortURLResponse))
			assert.Equal(t, test.want.code, getFullURLResponse.StatusCode())
			assert.True(t, containsShortURL(findByUserIDResponse, string(shortURLResponse.Body()), test.targetFullURL))
		})
	}
}

func TestGetWithCompress(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name          string
		targetFullURL string
		want          want
	}{
		{
			name:          "Successfully get the full body by the short body",
			targetFullURL: "https://yandex.ru",
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURLResponse, _ := createShortURLRequest(server.URL, test.targetFullURL).Send()

			getFullURLRequest := resty.New().SetRedirectPolicy(resty.NoRedirectPolicy()).R()
			getFullURLRequest.Method = http.MethodGet
			getFullURLRequest.SetHeader("Accept-Encoding", "compress")
			getFullURLRequest.SetHeader("Content-Type", "application/json")
			getFullURLRequest.URL = server.URL + extractHashKeyURLFrom(string(shortURLResponse.Body()))

			getFullURLResponse, _ := getFullURLRequest.Send()

			assert.NotEmpty(t, getTokenFromResponse(shortURLResponse))
			assert.Equal(t, test.want.code, getFullURLResponse.StatusCode())
			assert.Equal(t, test.targetFullURL, getFullURLResponse.Header().Get(LocationHeader))
		})
	}
}

func containsShortURL(slice []model.FindURLByUserIDResponse, shortURL, fullURL string) bool {
	for _, item := range slice {
		if item.ShortURL == shortURL && item.OriginalURL == fullURL {
			return true
		}
	}
	return false
}
