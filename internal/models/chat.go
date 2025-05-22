// internal/models/chat.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID                uuid.UUID `json:"id"`
	AccountID         uuid.UUID `json:"account_id"`
	Department        string    `json:"department"` // ex: financeiro, comercial
	Title             string    `json:"title"`
	Instructions      string    `json:"instructions"`
	PhoneNumber       string    `json:"phone_number"`
	EvolutionInstance string    `json:"evolution_instance"`
	WebhookURL        string    `json:"webhook_url"`
	Status            string    `json:"status"` // ativo, inativo
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
