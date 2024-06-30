package data

import "time"

type FileEntity struct {
	ID         int       `json:"id"`
	ExternalId string    `json:"external_id"`
	State      string    `json:"state"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
