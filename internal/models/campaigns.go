// File: /interna/models/campaigns.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Campaign representa uma campanha de marketing
type Campaign struct {
	ID          uuid.UUID       `json:"id"`
	AccountID   uuid.UUID       `json:"account_id"`
	Name        string          `json:"name"`
	Description *string         `json:"description,omitempty"`
	Channels    ChannelsConfig  `json:"channels"` // JSONB
	Filters     AudienceFilters `json:"filters"`  // JSONB
	Status      string          `json:"status"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

var ChannelType string

const (
	EmailChannel    = "email"
	WhatsappChannel = "whatsapp"
)

// Define the CampaignChannelsTypes
var CampaignChannelsTypes = []string{EmailChannel, WhatsappChannel}

// ChannelsConfig define a estrutura dos canais e templates usados (email, whatsapp, etc.)
type ChannelsConfig map[string]ChannelConfig

// ChannelConfig define um canal e o template associado
type ChannelConfig struct {
	TemplateID uuid.UUID `json:"template"`
	Priority   int       `json:"priority"`
}

// AudienceFilters define os critérios do público-alvo para a campanha
type AudienceFilters struct {
	Tags           []string `json:"tags,omitempty"`
	Gender         *string  `json:"gender,omitempty"`
	BirthDateRange struct {
		Start string `json:"start,omitempty"`
		End   string `json:"end,omitempty"`
	} `json:"birth_date_range,omitempty"`
}
