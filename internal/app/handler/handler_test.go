package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/repository/inmemory"
	"github.com/faust8888/shortener/internal/app/route"
	"github.com/go-resty/resty/v2"
	"net/http"
	"net/http/httptest"
	"net/url"
)

func startTestServer() *httptest.Server {
	cfg := config.Create()
	handler := Create(inmemory.NewRepository(), cfg.BaseShortURL)
	return httptest.NewServer(route.Create(handler))
}

func createShortURLRequest(url string, body interface{}, headers ...RequestHeader) *resty.Request {
	req := resty.New().R()
	req.Method = http.MethodPost
	req.URL = url
	req.Body = body
	for _, header := range headers {
		req.SetHeader(header.HeaderName, header.HeaderValue)
	}
	return req
}

func extractHashKeyURLFrom(shortURL string) string {
	parsedURL, _ := url.Parse(shortURL)
	return parsedURL.Path
}

type RequestHeader struct {
	HeaderName  string
	HeaderValue string
}
