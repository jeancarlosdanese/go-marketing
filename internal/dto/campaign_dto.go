// File: /internal/dto/campaign_dto.go

package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// CampaignCreateDTO define os dados para criar uma campanha
type CampaignCreateDTO struct {
	Name        string                  `json:"name"`
	Description *string                 `json:"description,omitempty"`
	Channels    models.ChannelsConfig   `json:"channels"`
	Filters     *models.AudienceFilters `json:"filters"`
}

// Validate valida os dados do CampaignCreateDTO
func (c *CampaignCreateDTO) Validate() error {
	if c.Name == "" {
		return errors.New("o nome da campanha é obrigatório")
	}
	if len(c.Channels) == 0 {
		return errors.New("pelo menos um canal deve ser definido")
	}
	for _, channel := range c.Channels {
		if channel.TemplateID == uuid.Nil {
			return errors.New("cada canal deve ter um template definido")
		}
		if channel.Priority < 1 {
			return errors.New("a prioridade do canal deve ser pelo menos 1")
		}
	}
	return nil
}

// CampaignUpdateDTO define os dados permitidos para atualização de uma campanha
type CampaignUpdateDTO struct {
	Name        *string                 `json:"name,omitempty"`
	Description *string                 `json:"description,omitempty"`
	Channels    *models.ChannelsConfig  `json:"channels,omitempty"`
	Filters     *models.AudienceFilters `json:"filters,omitempty"`
	Status      *models.CampaignStatus  `json:"status,omitempty"`
}

// Validate valida os dados do CampaignUpdateDTO
func (c *CampaignUpdateDTO) Validate() error {
	if c.Channels != nil && len(*c.Channels) == 0 {
		return errors.New("pelo menos um canal deve ser definido")
	}
	if c.Status != nil {
		validStatuses := map[models.CampaignStatus]bool{models.StatusPendente: true, models.StatusAtiva: true, models.StatusConcluida: true}
		if !validStatuses[*c.Status] {
			return errors.New("status inválido, deve ser 'pendente', 'ativa' ou 'concluida'")
		}
	}
	return nil
}

// CampaignUpdateStatusDTO define os dados permitidos para atualização de status da campanha
type CampaignUpdateStatusDTO struct {
	Status models.CampaignStatus `json:"status"`
}

// Validate valida os dados do CampaignUpdateStatusDTO
func (c *CampaignUpdateStatusDTO) Validate() error {
	validStatuses := map[models.CampaignStatus]bool{models.StatusPendente: true, models.StatusAtiva: true, models.StatusConcluida: true}
	if !validStatuses[c.Status] {
		return errors.New("status inválido, deve ser 'pendente', 'ativa' ou 'concluida'")
	}

	return nil
}

// CampaignResponseDTO estrutura a resposta para campanhas
type CampaignResponseDTO struct {
	ID          string                  `json:"id"`
	AccountID   string                  `json:"account_id"`
	Name        string                  `json:"name"`
	Description *string                 `json:"description,omitempty"`
	Channels    models.ChannelsConfig   `json:"channels"`
	Filters     *models.AudienceFilters `json:"filters,omitempty"`
	Status      models.CampaignStatus   `json:"status"`
	CreatedAt   string                  `json:"created_at"`
	UpdatedAt   string                  `json:"updated_at"`
}

// NewCampaignResponseDTO converte um modelo `Campaign` para um DTO de resposta
func NewCampaignResponseDTO(campaign *models.Campaign) CampaignResponseDTO {
	return CampaignResponseDTO{
		ID:          campaign.ID.String(),
		AccountID:   campaign.AccountID.String(),
		Name:        campaign.Name,
		Description: campaign.Description,
		Channels:    campaign.Channels,
		Filters:     campaign.Filters,
		Status:      campaign.Status,
		CreatedAt:   campaign.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   campaign.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
