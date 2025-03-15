package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/route"
	"github.com/faust8888/shortener/internal/app/storage/inmemory"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCreateShortURL(t *testing.T) {
	server := startTestServer()
	defer server.Close()

	type want struct {
		code           int
		responseRegexp string
		isError        bool
		errorMessage   string
	}
	tests := []struct {
		name string
		url  string
		want want
	}{
		{
			name: "Short URL successfully created for https://yandex.ru",
			url:  "https://yandex.ru",
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^https?://localhost:\d+/[\w/]+$`,
			},
		},
		{
			name: "Short URL successfully created for URL with full path",
			url:  "https://yandex.ru/path/fullPath",
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^https?://localhost:\d+/[\w/]+$`,
			},
		},
		{
			name: "Short URL wasn't created for invalid url",
			url:  "invalid_ru",
			want: want{
				code:         http.StatusBadRequest,
				isError:      true,
				errorMessage: "invalid url\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := createShortURLRequest(server.URL, test.url).Send()

			require.NoError(t, err)
			assert.Equal(t, test.want.code, resp.StatusCode())

			if !test.want.isError {
				assert.Regexp(t, test.want.responseRegexp, string(resp.Body()))
			} else {
				assert.Equal(t, test.want.errorMessage, string(resp.Body()))
			}
		})
	}
}

func TestGetFullURL(t *testing.T) {
	server := startTestServer()
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
			name:          "Successfully get the full url by the short url",
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

			assert.Equal(t, test.want.code, getFullURLResponse.StatusCode())
			assert.Equal(t, test.targetFullURL, getFullURLResponse.Header().Get(LocationHeader))
		})
	}
}

func startTestServer() *httptest.Server {
	cfg := config.Create()
	h := Create(inmemory.NewStorage(), cfg.BaseShortURL)
	return httptest.NewServer(route.Create(h))
}

func createShortURLRequest(url, body string) *resty.Request {
	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = url
	req.Body = body
	return req
}

func extractHashKeyURLFrom(shortURL string) string {
	parsedURL, _ := url.Parse(shortURL)
	return parsedURL.Path
}
