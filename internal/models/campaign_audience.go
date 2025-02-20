// File: /internal/models/campaign_audience.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// CampaignAudience representa um contato incluído em uma campanha específica
type CampaignAudience struct {
	ID         uuid.UUID               `json:"id"`
	CampaignID uuid.UUID               `json:"campaign_id"`
	ContactID  uuid.UUID               `json:"contact_id"`
	Type       string                  `json:"type"` // "email" ou "whatsapp"
	Status     string                  `json:"status"`
	MessageID  *string                 `json:"message_id,omitempty"`
	Feedback   *map[string]interface{} `json:"feedback_api,omitempty"` // JSONB
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}

// NewCampaignAudience cria uma nova instância de CampaignAudience
func NewCampaignAudience(campaignID, contactID uuid.UUID, messageType string) *CampaignAudience {
	return &CampaignAudience{
		CampaignID: campaignID,
		ContactID:  contactID,
		Type:       messageType,
		Status:     "pendente",
	}
}
