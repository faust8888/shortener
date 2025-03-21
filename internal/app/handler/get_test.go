package handler

import (
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGet(t *testing.T) {
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

			assert.Equal(t, test.want.code, getFullURLResponse.StatusCode())
			assert.Equal(t, test.targetFullURL, getFullURLResponse.Header().Get(LocationHeader))
		})
	}
}

func TestGetWithCompress(t *testing.T) {
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
			getFullURLRequest.SetHeader("Accept-Encoding", "gzip")
			getFullURLRequest.SetHeader("Content-Type", "application/json")
			getFullURLRequest.URL = server.URL + extractHashKeyURLFrom(string(shortURLResponse.Body()))

			getFullURLResponse, _ := getFullURLRequest.Send()

			assert.Equal(t, test.want.code, getFullURLResponse.StatusCode())
			assert.Equal(t, test.targetFullURL, getFullURLResponse.Header().Get(LocationHeader))
		})
	}
}
