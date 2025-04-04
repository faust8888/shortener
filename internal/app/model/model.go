package model

import "errors"

type CreateShortRequest struct {
	URL string `json:"url" validate:"required,url"`
}

func (req *CreateShortRequest) Validate() error {
	if req.URL == "" {
		return errors.New("url is required")
	}
	return nil
}

type CreateShortResponse struct {
	Result string `json:"result"`
}

type CreateShortRequestBatchItemRequest struct {
	CorrelationID string `json:"correlation_id" validate:"required,correlation_id"`
	OriginalURL   string `json:"original_url" validate:"required,original_url"`
}

type CreateShortRequestBatchItemResponse struct {
	CorrelationID string `json:"correlation_id" validate:"required,correlation_id"`
	ShortURL      string `json:"short_url" validate:"required,short_url"`
}

type CreateShortDTO struct {
	OriginalURL string
	ShortURL    string
	HashURL     string
}
