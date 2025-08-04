package repository

import "github.com/faust8888/shortener/internal/app/model"

// Repository — это интерфейс, определяющий основные операции над хранилищем коротких ссылок.
// Реализация может быть файловой, базой данных или в памяти.
type Repository interface {
	// Save сохраняет одну пару (hashURL -> fullURL) для указанного пользователя.
	//
	// Параметры:
	//   - urlHash: хэш-ключ для короткой ссылки.
	//   - fullURL: оригинальный URL.
	//   - userID: идентификатор пользователя.
	//
	// Возвращает:
	//   - error: nil, если успешно, иначе — ошибку.
	Save(urlHash, fullURL, userID string) error

	// FindByHash находит оригинальный URL по его хэш-ключу.
	//
	// Параметр:
	//   - hashURL: хэш-ключ короткой ссылки.
	//
	// Возвращает:
	//   - string: оригинальный URL.
	//   - error: nil, если найдено, иначе — ошибку.
	FindByHash(hashURL string) (string, error)

	// FindAllByUserID возвращает все короткие ссылки, принадлежащие пользователю.
	//
	// Параметр:
	//   - userID: идентификатор пользователя.
	//
	// Возвращает:
	//   - []model.FindURLByUserIDResponse: список ссылок пользователя.
	//   - error: nil, если успешно, иначе — ошибку.
	FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error)

	// SaveAll сохраняет несколько ссылок за один раз (пакетная операция).
	//
	// Параметры:
	//   - batch: карта хэшей и DTO с данными о ссылках.
	//   - userID: идентификатор пользователя.
	//
	// Возвращает:
	//   - error: nil, если успешно, иначе — ошибку.
	SaveAll(batch map[string]model.CreateShortDTO, userID string) error

	// DeleteAll асинхронно удаляет несколько коротких ссылок пользователя.
	//
	// Параметры:
	//   - shortURLs: список идентификаторов (хэшей) ссылок для удаления.
	//   - userID: идентификатор пользователя.
	//
	// Возвращает:
	//   - error: nil, если запрос на удаление принят, иначе — ошибку.
	DeleteAll(shortURLs []string, userID string) error

	// Ping проверяет доступность хранилища.
	//
	// Возвращает:
	//   - bool: true, если хранилище доступно.
	//   - error: nil, если хранилище доступно, иначе — ошибку.
	Ping() (bool, error)

	// Collect собирает статистику по уникальным пользователям и URL из хранилища.
	//
	// Возвращает:
	//   - *model.Statistic: структура со статистическими данными по URL и пользователям.
	//   - error: nil, если запрос выполнен успешно, иначе — ошибка.
	Collect() (*model.Statistic, error)
}
