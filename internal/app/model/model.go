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
