// File: /internal/models/campaign_message.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type CampaignMessage struct {
	ID          uuid.UUID  `json:"id"`
	CampaignID  uuid.UUID  `json:"campaign_id"`
	ContactID   *uuid.UUID `json:"contact_id,omitempty"`
	Channel     string     `json:"channel"` // "email" ou "whatsapp"
	Saudacao    string     `json:"saudacao"`
	Corpo       string     `json:"corpo"`
	Finalizacao string     `json:"finalizacao"`
	Assinatura  string     `json:"assinatura"`
	PromptUsado string     `json:"prompt_usado"`
	Feedback    []string   `json:"feedback"` // armazenado como JSONB
	Version     int        `json:"version"`
	IsApproved  bool       `json:"is_approved"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
