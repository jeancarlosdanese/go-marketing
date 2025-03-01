// File: /internal/dto/campaign_audience_dto.go

package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// CampaignAudienceCreateDTO define os dados para adicionar contatos à audiência de uma campanha
type CampaignAudienceCreateDTO struct {
	ContactIDs []uuid.UUID         `json:"contact_ids"`
	Type       *models.ChannelType `json:"type"`
}

// CampaignAudienceCreateByFilterDTO define os dados para filtrar a audiência de uma campanha
type CampaignAudienceCreateByFilterDTO struct {
	// "name", "email", "whatsapp", "cidade", "estado", "bairro", "gender", "birth_date_start", "birth_date_end", "last_contact_at", "interesses", "perfil", "eventos", "tags"
	Filters     *map[string]string `json:"filters"`
	CurrentPage int                `json:"current_page"` // Página atual
	PerPage     int                `json:"per_page"`     // Registros por página
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

// Validate valida os dados do CampaignAudienceCreateByFilterDTO
func (c *CampaignAudienceCreateByFilterDTO) Validate() error {
	if c.CurrentPage <= 1 {
		return errors.New("current_page deve ser maior que 0")
	}
	if c.PerPage <= 1 {
		return errors.New("per_page deve ser maior que 0")
	}
	return nil
}

// CampaignAudienceResponseDTO estrutura a resposta para um contato dentro de uma campanha
type CampaignAudienceResponseDTO struct {
	ID         string                  `json:"id"`
	CampaignID string                  `json:"campaign_id"`
	ContactID  string                  `json:"contact_id"`
	Type       models.ChannelType      `json:"type"`
	Status     models.CampaignStatus   `json:"status"`
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
