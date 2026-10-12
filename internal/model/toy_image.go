package model

import (
	"time"

	"github.com/google/uuid"
)

type ToyImage struct {
	ID        uuid.UUID `json:"id"`
	ToyID     uuid.UUID `json:"toy_id"`
	Url       string    `json:"url"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
