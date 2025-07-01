package handler

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	server := startTestServer(t)
	defer server.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "Successfully PingDatabase",
			want: want{
				code: http.StatusOK,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = fmt.Sprintf("%s/ping", server.URL)

			resp, err := req.Send()

			assert.NoError(t, err)
			assert.Equal(t, test.want.code, resp.StatusCode())
		})
	}
}
