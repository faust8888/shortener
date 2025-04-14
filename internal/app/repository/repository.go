package repository

import "github.com/faust8888/shortener/internal/app/model"

type Repository interface {
	Save(urlHash string, fullURL string) error
	FindByHash(hashURL string) (string, error)
	SaveAll(batch map[string]model.CreateShortDTO) error
	Ping() (bool, error)
}
