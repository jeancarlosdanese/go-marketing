// File: /internal/dto/campaign_audience_dto.go

package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// CampaignAudienceCreateDTO define os dados para adicionar contatos à audiência de uma campanha
type CampaignAudienceCreateDTO struct {
	ContactIDs []uuid.UUID `json:"contact_ids"`
	Type       *string     `json:"type"`
}

// Validate valida os dados do CampaignAudienceCreateDTO
func (c *CampaignAudienceCreateDTO) Validate() error {
	if len(c.ContactIDs) == 0 {
		return errors.New("deve haver pelo menos um contact_id")
	}
	if c.Type != nil && *c.Type != "email" && *c.Type != "whatsapp" {
		return errors.New("o tipo deve ser 'email' ou 'whatsapp'")
	}
	return nil
}

// CampaignAudienceResponseDTO estrutura a resposta para um contato dentro de uma campanha
type CampaignAudienceResponseDTO struct {
	ID         string                  `json:"id"`
	CampaignID string                  `json:"campaign_id"`
	ContactID  string                  `json:"contact_id"`
	Type       string                  `json:"type"`
	Status     string                  `json:"status"`
	MessageID  *string                 `json:"message_id,omitempty"`
	Feedback   *map[string]interface{} `json:"feedback_api,omitempty"`
	CreatedAt  string                  `json:"created_at"`
	UpdatedAt  string                  `json:"updated_at"`
}

// NewCampaignAudienceResponseDTO converte um modelo `CampaignAudience` para um DTO de resposta
func NewCampaignAudienceResponseDTO(audience *models.CampaignAudience) CampaignAudienceResponseDTO {
	return CampaignAudienceResponseDTO{
		ID:         audience.ID.String(),
		CampaignID: audience.CampaignID.String(),
		ContactID:  audience.ContactID.String(),
		Type:       audience.Type,
		Status:     audience.Status,
		MessageID:  audience.MessageID,
		Feedback:   audience.Feedback,
		CreatedAt:  audience.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  audience.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
