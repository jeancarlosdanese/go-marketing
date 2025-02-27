// File: /interna/models/campaigns.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Campaign representa uma campanha de marketing
type Campaign struct {
	ID          uuid.UUID        `json:"id"`
	AccountID   uuid.UUID        `json:"account_id"`
	Name        string           `json:"name"`
	Description *string          `json:"description,omitempty"`
	Channels    ChannelsConfig   `json:"channels"`
	Filters     *AudienceFilters `json:"filters,omitempty"`
	Status      CampaignStatus   `json:"status"` // Usa o enum CampaignStatus
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
}

// ChannelsConfig define a estrutura dos canais e templates usados (email, whatsapp, etc.)
type ChannelsConfig map[string]ChannelConfig

// ChannelConfig define um canal e o template associado
type ChannelConfig struct {
	TemplateID uuid.UUID `json:"template"`
	Priority   int       `json:"priority"`
}

type DateRange struct {
	Start *string `json:"start,omitempty"`
	End   *string `json:"end,omitempty"`
}

// AudienceFilters define os critérios do público-alvo para a campanha
type AudienceFilters struct {
	Tags           []*string  `json:"tags,omitempty"`
	Gender         *string    `json:"gender,omitempty"`
	BirthDateRange *DateRange `json:"birth_date_range,omitempty"`
}
