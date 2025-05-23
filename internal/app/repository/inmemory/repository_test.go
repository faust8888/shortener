package inmemory

import (
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInMemoryStorageFindByHashURLAndSave(t *testing.T) {
	tests := []struct {
		name                string
		urlHashForSaving    string
		urlHashForSearching string
		fullURL             string
		userID              string
		wantErr             bool
	}{
		{
			name:                "Successfully saved",
			urlHashForSaving:    "qwerty12345",
			urlHashForSearching: "qwerty12345",
			userID:              "12345",
			fullURL:             "https://yandex.ru",
			wantErr:             false,
		},
		{
			name:                "Not found full URL with the wrong hash URL",
			urlHashForSaving:    "qwerty12345",
			urlHashForSearching: "asdfgh6789",
			userID:              "12345",
			fullURL:             "https://yandex.ru",
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewInMemoryRepository(config.Create())

			s.Save(tt.urlHashForSaving, tt.fullURL, tt.userID)
			returnedFullURL, err := s.FindByHash(tt.urlHashForSearching)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.fullURL, returnedFullURL)
			}
		})
	}
}
