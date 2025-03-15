package inmemory

import (
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
		wantErr             bool
	}{
		{
			name:                "Successfully saved",
			urlHashForSaving:    "qwerty12345",
			urlHashForSearching: "qwerty12345",
			fullURL:             "https://yandex.ru",
			wantErr:             false,
		},
		{
			name:                "Not found full URL with the wrong hash URL",
			urlHashForSaving:    "qwerty12345",
			urlHashForSearching: "asdfgh6789",
			fullURL:             "https://yandex.ru",
			wantErr:             true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewStorage()

			s.Save(tt.urlHashForSaving, tt.fullURL)
			returnedFullURL, err := s.FindByHashURL(tt.urlHashForSearching)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.fullURL, returnedFullURL)
			}
		})
	}
}
