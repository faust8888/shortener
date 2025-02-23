package handlers

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateShortURL(t *testing.T) {
	h := CreateInMemoryHandler()
	type want struct {
		code           int
		responseRegexp string
		isError        bool
		errorMessage   string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "Short URL successfully created for https://yandex.ru",
			body: "https://yandex.ru",
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^https?://localhost:\d+/[\w/]+$`,
			},
		},
		{
			name: "Short URL successfully created for URL with full path",
			body: "https://yandex.ru/path/fullPath",
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^https?://localhost:\d+/[\w/]+$`,
			},
		},
		{
			name: "Short URL wasn't created for invalid url",
			body: "invalid_ru",
			want: want{
				code:         http.StatusBadRequest,
				isError:      true,
				errorMessage: "invalid url\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, statusCode := createShortURL(h, test.body)

			assert.Equal(t, test.want.code, statusCode)

			if !test.want.isError {
				assert.Regexp(t, test.want.responseRegexp, body)
			} else {
				assert.Equal(t, test.want.errorMessage, body)
			}
		})
	}
}

func TestGetFullURL(t *testing.T) {
	h := CreateInMemoryHandler()
	type want struct {
		code         int
		isError      bool
		errorMessage string
	}
	tests := []struct {
		name          string
		targetFullURL string
		want          want
	}{
		{
			name:          "Successfully get the full url by the shorten url",
			targetFullURL: "https://yandex.ru",
			want: want{
				code: http.StatusTemporaryRedirect,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shortURL, _ := createShortURL(h, test.targetFullURL)

			returnedFullURL, statusCode := getFullURL(h, shortURL)

			assert.Equal(t, test.want.code, statusCode)
			assert.Equal(t, test.targetFullURL, returnedFullURL)
		})
	}
}

func createShortURL(handler *Handler, targetURL string) (string, int) {
	request := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(targetURL))
	w := httptest.NewRecorder()

	handler.CreateShortURL(w, request)

	res := w.Result()
	defer res.Body.Close()
	responseBody, _ := io.ReadAll(res.Body)
	return string(responseBody), res.StatusCode
}

func getFullURL(handler *Handler, shortURL string) (string, int) {
	req := httptest.NewRequest(http.MethodGet, string(shortURL), nil)
	w := httptest.NewRecorder()

	handler.GetFullURL(w, req)

	res := w.Result()
	defer res.Body.Close()
	return res.Header.Get(LocationHeader), res.StatusCode
}
