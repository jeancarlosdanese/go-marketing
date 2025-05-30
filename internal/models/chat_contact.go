// internal/models/chat_contact.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type ChatContact struct {
	ID                uuid.UUID `json:"id"`
	AccountID         uuid.UUID `json:"account_id"`
	ChatID            uuid.UUID `json:"chat_id"`
	ContactID         uuid.UUID `json:"contact_id"`
	WhatsappContactID uuid.UUID `json:"whatsapp_contact_id"`
	Status            string    `json:"status"` // aberto, pendente, fechado
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
