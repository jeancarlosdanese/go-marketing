// File: /internal/models/campaign_settings.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// CampaignSettings representa as configurações de envio de uma campanha
type CampaignSettings struct {
	ID                   uuid.UUID `json:"id"`
	CampaignID           uuid.UUID `json:"campaign_id"`
	Brand                string    `json:"brand"`
	Subject              string    `json:"subject"`
	Tone                 *string   `json:"tone,omitempty"`
	EmailFrom            string    `json:"email_from"`
	EmailReply           string    `json:"email_reply"`
	EmailFooter          *string   `json:"email_footer,omitempty"`
	EmailInstructions    string    `json:"email_instructions"`
	WhatsAppFrom         string    `json:"whatsapp_from"`
	WhatsAppReply        string    `json:"whatsapp_reply"`
	WhatsAppFooter       *string   `json:"whatsapp_footer,omitempty"`
	WhatsAppInstructions string    `json:"whatsapp_instructions"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
