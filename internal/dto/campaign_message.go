// FIle: /internal/dto/campaign_message.go

package dto

import (
	"time"

	"github.com/google/uuid"
)

// CampaignMessageDTO representa uma mensagem a ser enviada, unificando a campanha e o contato
type CampaignMessageDTO struct {
	ID         uuid.UUID `json:"id"`
	CampaignID uuid.UUID `json:"campaign_id"`
	ContactID  uuid.UUID `json:"contact_id"`
	Type       string    `json:"type"` // "email" ou "whatsapp"
	Status     string    `json:"status"`

	// Dados do contato
	Name          string                 `json:"name"`
	Email         *string                `json:"email,omitempty"`
	WhatsApp      *string                `json:"whatsapp,omitempty"`
	Gender        *string                `json:"gender,omitempty"`
	BirthDate     *time.Time             `json:"birth_date,omitempty"`
	Bairro        *string                `json:"bairro,omitempty"`
	Cidade        *string                `json:"cidade,omitempty"`
	Estado        *string                `json:"estado,omitempty"`
	Tags          map[string]interface{} `json:"tags"`
	History       *string                `json:"history,omitempty"`
	OptOutAt      *time.Time             `json:"opt_out_at,omitempty"`
	LastContactAt *time.Time             `json:"last_contact_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}
