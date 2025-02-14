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

// IsAdmin verifica se a conta é admin baseada no ID fixo
func (a *Account) IsAdmin() bool {
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	return a.ID == adminID
}
