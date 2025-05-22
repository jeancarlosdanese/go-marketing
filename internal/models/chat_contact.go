// internal/models/chat_contact.go

package models

import "time"

type ChatContact struct {
	ID        string    `json:"id"`
	AccountID string    `json:"account_id"`
	ChatID    string    `json:"chat_id"`
	ContactID string    `json:"contact_id"`
	Status    string    `json:"status"` // aberto, pendente, fechado
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
