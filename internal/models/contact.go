// File: /internal/models/contact.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Contact representa um contato armazenado no sistema
type Contact struct {
	ID        uuid.UUID  `json:"id"`
	AccountID uuid.UUID  `json:"account_id"`
	Name      string     `json:"name"`
	Email     string     `json:"email,omitempty"`
	WhatsApp  string     `json:"whatsapp,omitempty"`
	Gender    *string    `json:"gender,omitempty"`     // masculino, feminino ou outro (opcional).
	BirthDate *time.Time `json:"birth_date,omitempty"` // Data de nascimento (opcional).
	History   *string    `json:"history,omitempty"`    // Pequeno texto sobre o contato (opcional, pode ajudar na personalização da IA).
	OptOutAt  *time.Time `json:"opt_out_at,omitempty"` // Data de saída da lista, se o contato optou por sair da lista de marketing.
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
