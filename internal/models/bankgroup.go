package models

import (
	"time"

	"github.com/google/uuid"
)

type BankGroup struct {
	GroupID     uuid.UUID `json:"group_id" db:"group_id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description" db:"description"`
	LogoURL     *string   `json:"logo_url" db:"logo_url"`
	Website     *string   `json:"website" db:"website"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
