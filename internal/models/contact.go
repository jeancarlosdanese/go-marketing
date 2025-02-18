// File: /internal/models/contact.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Contact representa um contato no sistema
type Contact struct {
	ID            uuid.UUID   `json:"id"`
	AccountID     uuid.UUID   `json:"account_id"`
	Name          string      `json:"name"`
	Email         *string     `json:"email,omitempty"`
	WhatsApp      *string     `json:"whatsapp,omitempty"`
	Gender        *string     `json:"gender,omitempty"`
	BirthDate     *time.Time  `json:"birth_date,omitempty"`
	Bairro        *string     `json:"bairro,omitempty"`
	Cidade        *string     `json:"cidade,omitempty"`
	Estado        *string     `json:"estado,omitempty"`
	Tags          ContactTags `json:"tags,omitempty"` // JSONB estruturado
	History       *string     `json:"history,omitempty"`
	OptOutAt      *time.Time  `json:"opt_out_at,omitempty"`
	LastContactAt *time.Time  `json:"last_contact_at,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
	UpdatedAt     time.Time   `json:"updated_at"`
}

// ContactTags estrutura as tags como JSONB
type ContactTags struct {
	Interesses []*string `json:"interesses,omitempty"`
	Perfil     *string   `json:"perfil,omitempty"`
	Eventos    []*string `json:"eventos,omitempty"`
}
