// File: /internal/dto/campaign_settings_dto.go

package dto

import (
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// CampaignSettingsDTO representa as configurações de envio de uma campanha
type CampaignSettingsDTO struct {
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
}

// Valida os dados da CampaignSettingsDTO antes de persistir
func (c *CampaignSettingsDTO) Validate() error {
	// 1. Validação de campos obrigatórios
	if c.CampaignID == uuid.Nil {
		return errors.New("campaign_id é obrigatório")
	}
	if c.Brand == "" || len(c.Brand) > 100 {
		return errors.New("brand é obrigatório e deve ter no máximo 100 caracteres")
	}
	if c.Subject == "" || len(c.Subject) > 150 {
		return errors.New("subject é obrigatório e deve ter no máximo 150 caracteres")
	}
	if c.EmailFrom == "" || len(c.EmailFrom) > 150 {
		return errors.New("email_from é obrigatório e deve ter no máximo 150 caracteres")
	}
	if c.EmailReply == "" || len(c.EmailReply) > 150 {
		return errors.New("email_reply é obrigatório e deve ter no máximo 150 caracteres")
	}
	if c.EmailInstructions == "" {
		return errors.New("email_instructions é obrigatório")
	}
	if c.WhatsAppFrom == "" || len(c.WhatsAppFrom) > 20 {
		return errors.New("whatsapp_from é obrigatório e deve ter no máximo 20 caracteres")
	}
	if c.WhatsAppReply == "" || len(c.WhatsAppReply) > 20 {
		return errors.New("whatsapp_reply é obrigatório e deve ter no máximo 20 caracteres")
	}
	if c.WhatsAppInstructions == "" {
		return errors.New("whatsapp_instructions é obrigatório")
	}

	// 2. Validação de e-mails
	if err := utils.ValidateEmail(c.EmailFrom); err != nil {
		return errors.New("email_from inválido")
	}
	if err := utils.ValidateEmail(c.EmailReply); err != nil {
		return errors.New("email_reply inválido")
	}

	// 3. Validação de WhatsApp (apenas números, com prefixo internacional opcional)
	if err := utils.ValidateWhatsApp(c.WhatsAppFrom); err != nil {
		return errors.New("whatsapp_from inválido")
	}
	if err := utils.ValidateWhatsApp(c.WhatsAppReply); err != nil {
		return errors.New("whatsapp_reply inválido")
	}

	// 4. Validação do tom de voz (Tone)
	if c.Tone != nil {
		validTones := map[string]bool{"formal": true, "casual": true, "neutro": true}
		if !validTones[strings.ToLower(*c.Tone)] {
			return errors.New("tone deve ser 'formal', 'casual' ou 'neutro'")
		}
	}

	return nil
}

// Normalize normaliza os dados da CampaignSettingsDTO antes de persistir
func (c *CampaignSettingsDTO) Normalize() {
	c.Brand = strings.TrimSpace(c.Brand)
	c.Subject = strings.TrimSpace(c.Subject)
	c.EmailFrom = *utils.NormalizeEmail(&c.EmailFrom)
	c.EmailReply = *utils.NormalizeEmail(&c.EmailReply)
	c.EmailInstructions = strings.TrimSpace(c.EmailInstructions)
	c.WhatsAppFrom = *utils.FormatWhatsAppOnlyNumbers(&c.WhatsAppFrom)
	c.WhatsAppReply = *utils.FormatWhatsAppOnlyNumbers(&c.WhatsAppReply)
	c.WhatsAppInstructions = strings.TrimSpace(c.WhatsAppInstructions)
}

func (c *CampaignSettingsDTO) ToModel() models.CampaignSettings {
	c.Normalize()

	settings := models.CampaignSettings{
		CampaignID:           c.CampaignID,
		Brand:                c.Brand,
		Subject:              c.Subject,
		Tone:                 c.Tone,
		EmailFrom:            c.EmailFrom,
		EmailReply:           c.EmailReply,
		EmailFooter:          c.EmailFooter,
		EmailInstructions:    c.EmailInstructions,
		WhatsAppFrom:         c.WhatsAppFrom,
		WhatsAppReply:        c.WhatsAppReply,
		WhatsAppFooter:       c.WhatsAppFooter,
		WhatsAppInstructions: c.WhatsAppInstructions,
	}

	return settings
}
