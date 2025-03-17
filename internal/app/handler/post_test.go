package handler

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestPost(t *testing.T) {
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
			name: "Short URL wasn't created for invalid body",
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

func TestPostWithJson(t *testing.T) {
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
		body string
		want want
	}{
		{
			name: "Short URL successfully created for https://yandex.ru",
			body: `{"url":"https://yandex.ru"}`,
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^\{\"result\":\"http:\/\/localhost:\d{1,5}\/[a-zA-Z0-9\-_\/]+\"\}$`,
			},
		},
		{
			name: "Short URL successfully created for URL with full path",
			body: `{"url":"https://yandex.ru/path/fullPath"}`,
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^\{\"result\":\"http:\/\/localhost:\d{1,5}\/[a-zA-Z0-9\-_\/]+\"\}$`,
			},
		},
		{
			name: "Short URL wasn't created for invalid body",
			body: `{"url_wrong_name":"https://yandex.ru"}`,
			want: want{
				code:         http.StatusBadRequest,
				isError:      true,
				errorMessage: "url is required\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := createShortURLRequest(server.URL+"/api/shorten", test.body).Send()

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
