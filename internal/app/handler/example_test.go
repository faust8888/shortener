package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/faust8888/shortener/internal/app/repository/inmemory"
	"github.com/faust8888/shortener/internal/app/service"
	"net/http"
	"net/http/httptest"
)

// mockRepo — минимальная реализация repository.Repository для тестов.
type mockRepo struct{}

func (m *mockRepo) Save(urlHash, fullURL, userID string) error {
	return nil
}

func (m *mockRepo) FindByHash(hashURL string) (string, error) {
	if hashURL == "abc123" {
		return "http://example.com", nil
	}
	return "", fmt.Errorf("not found")
}

func (m *mockRepo) FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error) {
	if userID == "testuser123" {
		return []model.FindURLByUserIDResponse{
			{ShortURL: "http://your-shortener.com/abc123", OriginalURL: "http://example.com"},
		}, nil
	}
	return nil, nil
}

func (m *mockRepo) SaveAll(batch map[string]model.CreateShortDTO, userID string) error {
	return nil
}

func (m *mockRepo) DeleteAll(shortURLs []string, userID string) error {
	return nil
}

func (m *mockRepo) Ping() (bool, error) {
	return true, nil
}

// batchSaverMock — реализация интерфейса batchSaver для тестов.
type batchSaverMock struct{}

func (b *batchSaverMock) CreateWithBatch(batch []model.CreateShortRequestBatchItemRequest, userID string) ([]model.CreateShortRequestBatchItemResponse, error) {
	responses := make([]model.CreateShortRequestBatchItemResponse, 0, len(batch))
	for _, item := range batch {
		responses = append(responses, model.CreateShortRequestBatchItemResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      fmt.Sprintf("http://your-shortener.com/%s", item.CorrelationID[0:3]),
		})
	}
	return responses, nil
}

// creatorMock — реализация интерфейса creator для тестов.
type creatorMock struct{}

func (c *creatorMock) Create(fullURL string, userID string) (string, error) {
	return fmt.Sprintf("http://your-shortener.com/%s", fullURL[len(fullURL)-3:]), nil
}

// pingCheckerMock — реализация интерфейса PingChecker для тестов.
type pingCheckerMock struct{}

func (p *pingCheckerMock) Ping() (bool, error) {
	return true, nil
}

// deleterMock — реализация интерфейса deleter для тестов.
type deleterMock struct{}

func (d *deleterMock) DeleteAsync(ids []string, userID string) error {
	return nil
}

// createTestHandler создаёт готовый Handler для тестирования.
func createTestHandler() *Handler {
	cfg := config.Create()
	shortener := service.CreateShortener(inmemory.NewInMemoryRepository(cfg), cfg.BaseShortURL)
	return CreateHandler(shortener, &pingCheckerMock{}, cfg)
}

// ExampleCreateWithBatch демонстрирует использование эндпоинта /api/shorten/Batch.
func ExampleBatch_CreateLinkWithBatch() {
	h := createTestHandler()

	requestBody := []model.CreateShortRequestBatchItemRequest{
		{CorrelationID: "id1", OriginalURL: "http://example.com/1"},
		{CorrelationID: "id2", OriginalURL: "http://example.com/2"},
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten/Batch", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.Batch.CreateLinkWithBatch(rec, req)

	result := rec.Result()
	defer result.Body.Close()

	fmt.Println("Status Code:", result.StatusCode)

	// Output:
	// Status Code: 201
}

// ExampleCreateWithJSON демонстрирует использование эндпоинта /api/shorten.
func ExampleCreateWithJSON() {
	h := createTestHandler()

	requestBody := model.CreateShortRequest{
		URL: "http://example.com",
	}

	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	h.CreateWithJSON.CreateLinkWithJSON(rec, req)

	result := rec.Result()
	defer result.Body.Close()

	fmt.Println("Status Code:", result.StatusCode)

	// Output:
	// Status Code: 201
}

// ExampleCreate демонстрирует использование эндпоинта /.
func ExampleCreate() {
	h := createTestHandler()

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString("http://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	rec := httptest.NewRecorder()
	h.Create.CreateLink(rec, req)

	result := rec.Result()
	defer result.Body.Close()

	fmt.Println("Status Code:", result.StatusCode)

	// Output:
	// Status Code: 201
}
