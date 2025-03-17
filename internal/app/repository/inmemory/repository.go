package inmemory

import "fmt"

type Repository struct {
	bucket map[string]string
}

func (r *Repository) Save(urlHash string, fullURL string) {
	r.bucket[urlHash] = fullURL
}

func (r *Repository) FindByHash(hashURL string) (string, error) {
	if fullURL, exists := r.bucket[hashURL]; exists {
		return fullURL, nil
	}
	return "", fmt.Errorf("short url not found for %s", hashURL)
}

func NewRepository() *Repository {
	return &Repository{bucket: make(map[string]string)}
}
