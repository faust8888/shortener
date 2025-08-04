package model

import "errors"

// CreateShortRequest — это модель запроса на создание короткой ссылки через JSON.
//
// Используется в хендлере `createWithJSON`.
// Поле URL обязательно и должно быть корректным URL.
type CreateShortRequest struct {
	URL string `json:"url" validate:"required,url"`
}

// Validate проверяет, что поле URL не пустое.
//
// Возвращает:
//   - error: nil, если валидация успешна,
//     иначе — ошибку с описанием проблемы.
func (req *CreateShortRequest) Validate() error {
	if req.URL == "" {
		return errors.New("url is required")
	}
	return nil
}

// CreateShortResponse — это модель ответа при успешном создании короткой ссылки.
//
// Содержит поле Result — готовая короткая ссылка.
type CreateShortResponse struct {
	Result string `json:"result"`
}

// CreateShortRequestBatchItemRequest — элемент запроса для пакетного создания коротких ссылок.
//
// Содержит:
//   - CorrelationID: уникальный идентификатор элемента запроса,
//   - OriginalURL: оригинальный URL, который нужно сократить.
type CreateShortRequestBatchItemRequest struct {
	CorrelationID string `json:"correlation_id" validate:"required,correlation_id"`
	OriginalURL   string `json:"original_url" validate:"required,original_url"`
}

// CreateShortRequestBatchItemResponse — элемент ответа при пакетном создании коротких ссылок.
//
// Содержит:
//   - CorrelationID: идентификатор из запроса,
//   - ShortURL: созданная короткая ссылка.
type CreateShortRequestBatchItemResponse struct {
	CorrelationID string `json:"correlation_id" validate:"required,correlation_id"`
	ShortURL      string `json:"short_url" validate:"required,short_url"`
}

// FindURLByUserIDResponse — модель ответа при получении всех ссылок пользователя.
//
// Содержит:
//   - ShortURL: короткий URL,
//   - OriginalURL: оригинальный URL.
type FindURLByUserIDResponse struct {
	ShortURL    string `json:"short_url" validate:"required,short_url"`
	OriginalURL string `json:"original_url" validate:"required,original_url"`
}

// CreateShortDTO — это DTO (Data Transfer Object), используемый сервисом и репозиторием.
//
// Представляет данные для создания короткой ссылки вне зависимости от источника запроса.
//
// Содержит:
//   - OriginalURL: оригинальный URL.
//   - ShortURL: сгенерированный короткий URL.
//   - HashURL: хэш, использованный для генерации короткой ссылки.
type CreateShortDTO struct {
	OriginalURL string
	ShortURL    string
	HashURL     string
}

// Statistic — модель статистики по сокращённым URL.
// Содержит:
//   - Urls: Общее количество уникальных URL.
//   - Users: Общее количество уникальных пользователей.
type Statistic struct {
	Urls  int `json:"urls" validate:"required,urls"`
	Users int `json:"users" validate:"required,users"`
}
