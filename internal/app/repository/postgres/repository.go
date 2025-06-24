package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/faust8888/shortener/internal/app/config"
	"github.com/faust8888/shortener/internal/app/model"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/lib/pq"
	"time"
)

// ErrUniqueIndexConstraint — ошибка, возникающая при попытке дублирования записи по full_url.
var ErrUniqueIndexConstraint = errors.New("full_url unique index constraint violation")

// ErrRecordWasMarkedAsDeleted — ошибка, указывающая, что запись была помечена как удалённая.
var ErrRecordWasMarkedAsDeleted = errors.New("is_deleted is true for the record")

// Repository — реализация repository.Repository на основе PostgreSQL.
// Используется для хранения, поиска и удаления коротких ссылок в БД.
type Repository struct {
	db           *sql.DB // Подключение к базе данных
	baseShortURL string  // Базовый URL для формирования полного адреса
}

// Save сохраняет одну пару (hashURL -> fullURL) для указанного пользователя.
//
// Если запись уже существует — возвращает ErrUniqueIndexConstraint.
//
// Параметры:
//   - urlHash: хэш-ключ для короткой ссылки.
//   - fullURL: оригинальный URL.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: nil, если успешно, иначе — ошибку.
func (r *Repository) Save(urlHash string, fullURL string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	res, err := r.db.ExecContext(ctx,
		"INSERT INTO shortener (short_url, full_url, user_id) VALUES ($1, $2, $3) ON CONFLICT (full_url) DO NOTHING", urlHash, fullURL, userID)
	if err != nil {
		return fmt.Errorf("repository.postgres.save: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrUniqueIndexConstraint
	}
	return nil
}

// FindByHash находит оригинальный URL по его хэш-ключу.
//
// Также проверяет флаг is_deleted — если он установлен, возвращает ErrRecordWasMarkedAsDeleted.
//
// Параметр:
//   - hash: хэш-ключ короткой ссылки.
//
// Возвращает:
//   - string: оригинальный URL.
//   - error: nil, если найдено и не удалено, иначе — соответствующую ошибку.
func (r *Repository) FindByHash(hash string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := `
        SELECT full_url, is_deleted
        FROM shortener
        WHERE short_url = $1
    `
	var fullURL string
	var isDeleted bool
	err := r.db.QueryRowContext(ctx, query, hash).Scan(&fullURL, &isDeleted)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("short url not found for %s", hash)
		}
		return "", fmt.Errorf("failed to find short url by hash: %v", err)
	}
	if isDeleted {
		return "", ErrRecordWasMarkedAsDeleted
	}
	return fullURL, nil
}

// FindAllByUserID возвращает все короткие ссылки, принадлежащие пользователю.
//
// Формирует полные URL на основе baseShortURL.
//
// Параметр:
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - []model.FindURLByUserIDResponse: список ссылок пользователя.
//   - error: nil, если успешно, иначе — ошибку.
func (r *Repository) FindAllByUserID(userID string) ([]model.FindURLByUserIDResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := `
        SELECT full_url, short_url
        FROM shortener
        WHERE user_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres.repository.FindURLsByUserID: %w", err)
	}
	defer rows.Close()

	var results []model.FindURLByUserIDResponse
	for rows.Next() {
		var resp model.FindURLByUserIDResponse
		var shortURLWithoutBase string
		if err = rows.Scan(&resp.OriginalURL, &shortURLWithoutBase); err != nil {
			return nil, fmt.Errorf("postgres.repository.FindURLsByUserID: failed to scan row: %w", err)
		}
		resp.ShortURL = fmt.Sprintf("%s/%s", r.baseShortURL, shortURLWithoutBase)
		results = append(results, resp)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres.repository.FindURLsByUserID: error during row iteration: %w", err)
	}
	return results, nil
}

// SaveAll сохраняет несколько ссылок за один раз (пакетная операция).
// Выполняется в транзакции: если одна операция провалится — всё откатывается.
//
// Параметры:
//   - batch: карта хэшей и DTO с данными о ссылках.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: nil, если успешно, иначе — ошибку.
func (r *Repository) SaveAll(batch map[string]model.CreateShortDTO, userID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("postgres.repository.saveAll.begin - %w", err)
	}
	for _, batchItem := range batch {
		_, err = tx.ExecContext(context.Background(),
			"INSERT INTO shortener (short_url, full_url, user_id) VALUES ($1, $2, $3)", batchItem.HashURL, batchItem.OriginalURL, userID)
		if err != nil {
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				return fmt.Errorf("postgres.repository.saveAll.rollback - %w", rollbackErr)
			}
			return fmt.Errorf("postgres.repository.saveAll.insert: %w", err)
		}
	}
	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("postgres.repository.saveAll.commit - %w", err)
	}
	return nil
}

// DeleteAll асинхронно удаляет несколько коротких ссылок пользователя.
// Метит их как удалённые (is_deleted = true).
//
// Параметры:
//   - shortURLs: список идентификаторов (хэшей) ссылок для удаления.
//   - userID: идентификатор пользователя.
//
// Возвращает:
//   - error: nil, если запрос успешен, иначе — ошибку.
func (r *Repository) DeleteAll(shortURLs []string, userID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	query := `
        UPDATE shortener SET is_deleted = true 
        WHERE user_id = $1 AND short_url = ANY($2::text[])
    `
	rows, err := r.db.QueryContext(ctx, query, userID, pq.Array(shortURLs))
	if err != nil {
		return fmt.Errorf("postgres.repository.DeleteAll: %w", err)
	}
	if err = rows.Err(); err != nil {
		return fmt.Errorf("postgres.repository.DeleteAll: error during row iteration: %w", err)
	}
	defer rows.Close()
	return nil
}

// Ping проверяет доступность хранилища.
//
// Возвращает:
//   - bool: true, если база доступна.
//   - error: nil, если всё в порядке, иначе — ошибку.
func (r *Repository) Ping() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := r.db.PingContext(ctx); err != nil {
		return false, fmt.Errorf("couldn't ping the PostgreSQL server: %s", err.Error())
	}
	return true, nil
}

// NewPostgresRepository создаёт новый экземпляр Repository, подключаясь к PostgreSQL.
//
// Паникует, если не может установить соединение.
//
// Параметр:
//   - cfg: конфигурация приложения.
//
// Возвращает:
//   - *Repository: готовый к использованию объект репозитория.
func NewPostgresRepository(cfg *config.Config) *Repository {
	db, err := sql.Open("pgx", cfg.DataSourceName)
	if err != nil {
		panic(err)
	}
	return &Repository{
		db:           db,
		baseShortURL: cfg.BaseShortURL,
	}
}
