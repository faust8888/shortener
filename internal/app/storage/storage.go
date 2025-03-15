package storage

type Storage interface {
	Save(urlHash string, fullURL string)
	FindByHashURL(hashURL string) (string, error)
}
