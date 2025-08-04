package handler

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/mocks"
	"github.com/faust8888/shortener/internal/app/repository/inmemory"
	"github.com/faust8888/shortener/internal/app/route"
	"github.com/faust8888/shortener/internal/app/security"
	"github.com/faust8888/shortener/internal/app/service"
	"github.com/go-resty/resty/v2"
	"github.com/golang/mock/gomock"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func startTestServer(t *testing.T) *httptest.Server {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRepository := createRepositoryMock(ctrl)

	cfg := config.Create()
	shortener := service.CreateShortener(inmemory.NewInMemoryRepository(cfg), cfg.BaseShortURL)
	handler := CreateHandler(shortener, mockRepository, cfg)

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

func createRepositoryMock(ctrl *gomock.Controller) *mocks.MockRepository {
	mockRepository := mocks.NewMockRepository(ctrl)
	mockRepository.EXPECT().Ping().AnyTimes().Return(true, nil)
	return mockRepository
}

func getTokenFromResponse(res *resty.Response) string {
	for _, cookie := range res.Cookies() {
		if cookie.Name == security.AuthorizationTokenName {
			return cookie.Value
		}
	}
	return ""
}

type RequestHeader struct {
	HeaderName  string
	HeaderValue string
}
