package model

import (
	"time"

	"github.com/google/uuid"
)

type Room struct {
	ID        uuid.UUID `json:"id"         db:"id"`
	Name      string    `json:"name"       db:"name"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateRoomRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}
