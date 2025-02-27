package models

import (
	"time"

	"github.com/google/uuid"
)

// Template representa um modelo de mensagem para campanhas
type Template struct {
	ID          uuid.UUID   `json:"id"`
	AccountID   uuid.UUID   `json:"account_id"`
	Name        string      `json:"name"`
	Description *string     `json:"description,omitempty"`
	Channel     ChannelType `json:"channel"` // Usa o enum definido em `constants.go`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}
