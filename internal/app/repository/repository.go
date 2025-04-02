package repository

type Repository interface {
	Save(urlHash string, fullURL string) error
	FindByHash(hashURL string) (string, error)
	Ping() (bool, error)
}
