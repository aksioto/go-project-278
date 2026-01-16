package dto

import "time"

type LinkResponse struct {
	ID          int64     `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortName   string    `json:"short_name"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,rfc3986url"`
	ShortName   string `json:"short_name,omitempty" binding:"omitempty,min=3,max=32"`
}

type UpdateLinkRequest struct {
	OriginalURL string `json:"original_url" binding:"required,rfc3986url"`
	ShortName   string `json:"short_name" binding:"required,min=3,max=32"`
}
