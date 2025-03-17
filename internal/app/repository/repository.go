package repository

type Repository interface {
	Save(urlHash string, fullURL string)
	FindByHash(hashURL string) (string, error)
}
