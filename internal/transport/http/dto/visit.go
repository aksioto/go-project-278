package dto

import "time"

type VisitResponse struct {
	ID        int64     `json:"id"`
	LinkID    int64     `json:"link_id"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"user_agent"`
	Referer   string    `json:"referer,omitempty"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
