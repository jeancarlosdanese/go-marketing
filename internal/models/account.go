// File: /internal/models/account.go

package models

import (
	"github.com/google/uuid"
)

// Account representa a entidade de conta no banco de dados
type Account struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	WhatsApp string    `json:"whatsapp"`
}
