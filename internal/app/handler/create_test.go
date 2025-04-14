package handler

import (
	"bytes"
	"compress/gzip"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestPost(t *testing.T) {
	server := startTestServer(t)
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
				errorMessage: "hash for url: invalid url\n",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := createShortURLRequest(server.URL, test.url).Send()

			require.NoError(t, err)
			assert.Equal(t, test.want.code, resp.StatusCode())
			assert.NotEmpty(t, security.GetToken(resp.Cookies()))

			if !test.want.isError {
				assert.Regexp(t, test.want.responseRegexp, string(resp.Body()))
			} else {
				assert.Equal(t, test.want.errorMessage, string(resp.Body()))
			}
		})
	}
}

func TestPostWithJson(t *testing.T) {
	server := startTestServer(t)
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
			assert.NotEmpty(t, security.GetToken(resp.Cookies()))

			if !test.want.isError {
				assert.Regexp(t, test.want.responseRegexp, string(resp.Body()))
			} else {
				assert.Equal(t, test.want.errorMessage, string(resp.Body()))
			}
		})
	}
}

func TestPostWithJsonCompress(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	type want struct {
		code           int
		responseRegexp string
		isError        bool
		errorMessage   string
	}
	tests := []struct {
		name    string
		body    interface{}
		headers []RequestHeader
		want    want
	}{
		{
			name: "Short URL successfully created for https://yandex.ru",
			body: compressString(`{"url":"https://yandex.ru"}`),
			headers: []RequestHeader{
				{"Content-Type", "application/json"},
				{"Content-Encoding", "gzip"},
			},
			want: want{
				code:           http.StatusCreated,
				responseRegexp: `^\{\"result\":\"http:\/\/localhost:\d{1,5}\/[a-zA-Z0-9\-_\/]+\"\}$`,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := createShortURLRequest(server.URL+"/api/shorten", test.body, test.headers...).Send()

			require.NoError(t, err)
			assert.Equal(t, test.want.code, resp.StatusCode())
			assert.NotEmpty(t, security.GetToken(resp.Cookies()))

			if !test.want.isError {
				assert.Regexp(t, test.want.responseRegexp, string(resp.Body()))
			} else {
				assert.Equal(t, test.want.errorMessage, string(resp.Body()))
			}
		})
	}
}

func compressString(input string) []byte {
	// Create a buffer to hold the compressed data
	var buf bytes.Buffer

	// Create a new Gzip writer
	gz := gzip.NewWriter(&buf)

	// Write the input string to the Gzip writer
	if _, err := gz.Write([]byte(input)); err != nil {
		return nil
	}

	// Close the Gzip writer to finalize the compression
	if err := gz.Close(); err != nil {
		return nil
	}

	// Return the compressed data as a byte slice
	return buf.Bytes()
}
