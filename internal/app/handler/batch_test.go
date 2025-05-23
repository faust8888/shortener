package handler

import (
	"encoding/json"
	"github.com/faust8888/shortener/internal/app/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestCreateWithBatch(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name  string
		batch []model.CreateShortRequestBatchItemRequest
		want  want
	}{
		{
			name: "Successfully created Short URLs with batch",
			batch: []model.CreateShortRequestBatchItemRequest{
				{
					CorrelationID: "1",
					OriginalURL:   "https://yandex.ru",
				},
				{
					CorrelationID: "2",
					OriginalURL:   "https://google.com",
				},
			},
			want: want{
				code: http.StatusCreated,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			batchAsJSON, _ := json.Marshal(&test.batch)
			resp, err := createShortURLRequest(server.URL+"/api/shorten/batch", batchAsJSON).Send()

			require.NoError(t, err)
			assert.Equal(t, test.want.code, resp.StatusCode())
			assert.NotEmpty(t, getTokenFromResponse(resp))

			var batchResponse []model.CreateShortRequestBatchItemResponse
			_ = json.Unmarshal(resp.Body(), &batchResponse)

			assert.Equal(t, len(test.batch), len(batchResponse))
		})
	}
}
