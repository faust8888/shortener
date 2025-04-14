package repository

import "github.com/faust8888/shortener/internal/app/model"

type Repository interface {
	Save(urlHash, fullURL, userID string) error
	FindByHash(hashURL string) (string, error)
	FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error)
	SaveAll(batch map[string]model.CreateShortDTO, userID string) error
	Ping() (bool, error)
}
