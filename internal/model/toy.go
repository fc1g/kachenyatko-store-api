package model

import (
	"time"

	"github.com/google/uuid"
)

type Toy struct {
	ID          uuid.UUID `json:"id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Size        string    `json:"size"`
	Materials   string    `json:"materials"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
