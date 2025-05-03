// File: internal/dto/campaign_message_dto.go

package dto

import (
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type EmailData struct {
	Saudacao    template.HTML
	Corpo       template.HTML
	Finalizacao template.HTML
	Assinatura  template.HTML
}

// CampaignMessageDTO representa uma mensagem a ser enviada
type CampaignMessageDTO struct {
	ID         uuid.UUID `json:"id"`          // Audience ID
	AccountID  uuid.UUID `json:"account_id"`  // Account ID
	CampaignID uuid.UUID `json:"campaign_id"` // Campaign ID
	ContactID  uuid.UUID `json:"contact_id"`  // Contact ID
	Type       string    `json:"type"`        // "email" ou "whatsapp"
}

// CampaignAudienceDTO representa uma mensagem a ser enviada, unificando a campanha e o contato
type CampaignAudienceDTO struct {
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

// CampaignMessageFullDTO representa uma mensagem a ser enviada, unificando a campanha e o contato
type CampaignMessageFullDTO struct {
	ID         uuid.UUID `json:"id"`
	AccountID  uuid.UUID `json:"account_id"`
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
	Idade         *int                   `json:"idade,omitempty"`
	Bairro        *string                `json:"bairro,omitempty"`
	Cidade        *string                `json:"cidade,omitempty"`
	Estado        *string                `json:"estado,omitempty"`
	Tags          map[string]interface{} `json:"tags"`
	History       *string                `json:"history,omitempty"`
	OptOutAt      *time.Time             `json:"opt_out_at,omitempty"`
	LastContactAt *time.Time             `json:"last_contact_at,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`

	// Campos da campanha
	CampaignName         string     `json:"campaign_name"`                   // Nome da campanha
	CampaignDescription  *string    `json:"campaign_description,omitempty"`  // Descrição da campanha
	TemplateID           *uuid.UUID `json:"template_id,omitempty"`           // ID do Template
	Prioridade           *int       `json:"prioridade,omitempty"`            // Prioridade do canal
	Brand                string     `json:"brand"`                           // Nome da empresa/marca
	Subject              *string    `json:"subject,omitempty"`               // Assunto do e-mail
	Tone                 *string    `json:"tone,omitempty"`                  // Tom de voz (formal, informal, etc.)
	EmailFrom            *string    `json:"email_from,omitempty"`            // E-mail remetente
	EmailReplyTo         *string    `json:"email_reply_to,omitempty"`        // E-mail para resposta
	EmailFooter          *string    `json:"email_footer,omitempty"`          // Assinatura padrão do e-mail
	EmailInstructions    *string    `json:"email_instructions,omitempty"`    // Instruções para o e-mail
	WhatsAppFrom         *string    `json:"whatsapp_from,omitempty"`         // Nome do remetente no WhatsApp
	WhatsAppReplyTo      *string    `json:"whatsapp_reply_to,omitempty"`     // Número para resposta no WhatsApp
	WhatsAppFooter       *string    `json:"whatsapp_footer,omitempty"`       // Assinatura padrão do WhatsApp
	WhatsAppInstructions *string    `json:"whatsapp_instructions,omitempty"` // Instruções para o WhatsApp
}

// NewCCampaignMessageFullDTO cria um novo DTO e calcula a idade automaticamente
func NewCCampaignMessageFullDTO(data CampaignMessageFullDTO) CampaignMessageFullDTO {
	if data.BirthDate != nil {
		idade := calcularIdade(*data.BirthDate)
		data.Idade = &idade
	}
	return data
}

type CampaignMessageCreateDTO struct {
	CampaignID  uuid.UUID  `json:"campaign_id"`
	ContactID   *uuid.UUID `json:"contact_id,omitempty"`
	Channel     string     `json:"channel"` // "email" ou "whatsapp"
	Saudacao    string     `json:"saudacao"`
	Corpo       string     `json:"corpo"`
	Finalizacao string     `json:"finalizacao"`
	Assinatura  string     `json:"assinatura"`
	PromptUsado string     `json:"prompt_usado"`
	Feedback    []string   `json:"feedback,omitempty"`
	Version     int        `json:"version"`
	IsApproved  bool       `json:"is_approved"`
}

func (dto *CampaignMessageCreateDTO) ToModel() *models.CampaignMessage {
	return &models.CampaignMessage{
		CampaignID:  dto.CampaignID,
		ContactID:   dto.ContactID,
		Channel:     dto.Channel,
		Saudacao:    dto.Saudacao,
		Corpo:       dto.Corpo,
		Finalizacao: dto.Finalizacao,
		Assinatura:  dto.Assinatura,
		PromptUsado: dto.PromptUsado,
		Feedback:    dto.Feedback,
		Version:     dto.Version,
		IsApproved:  dto.IsApproved,
	}
}

func FromModel(msg *models.CampaignMessage) *CampaignMessageCreateDTO {
	return &CampaignMessageCreateDTO{
		CampaignID:  msg.CampaignID,
		ContactID:   msg.ContactID,
		Channel:     msg.Channel,
		Saudacao:    msg.Saudacao,
		Corpo:       msg.Corpo,
		Finalizacao: msg.Finalizacao,
		Assinatura:  msg.Assinatura,
		PromptUsado: msg.PromptUsado,
		Feedback:    msg.Feedback,
		Version:     msg.Version,
		IsApproved:  msg.IsApproved,
	}
}

type CampaignContentResult struct {
	Saudacao    string `json:"saudacao"`
	Corpo       string `json:"corpo"`
	Finalizacao string `json:"finalizacao"`
	Assinatura  string `json:"assinatura"`
}

type RenderedMessageDTO struct {
	Content    *CampaignContentResult `json:"content"`
	Preview    string                 `json:"preview"` // HTML renderizado ou markdown
	Prompt     string                 `json:"prompt"`
	Channel    string                 `json:"channel"`
	Version    int                    `json:"version"`
	IsApproved bool                   `json:"is_approved"`
}

func ToCampaignMessageFullDTO(
	account models.Account,
	campaign models.Campaign,
	settings models.CampaignSettings,
	contact models.Contact,
	channel string,
) CampaignMessageFullDTO {
	templateID := campaign.Channels[channel].TemplateID

	dto := CampaignMessageFullDTO{
		ID:            uuid.New(),
		AccountID:     account.ID,
		CampaignID:    campaign.ID,
		ContactID:     contact.ID,
		Type:          channel,
		Status:        "gerada",
		Name:          contact.Name,
		Email:         contact.Email,
		WhatsApp:      contact.WhatsApp,
		Gender:        contact.Gender,
		BirthDate:     contact.BirthDate,
		Bairro:        contact.Bairro,
		Cidade:        contact.Cidade,
		Estado:        contact.Estado,
		Tags:          models.ConvertContactTags(contact.Tags),
		History:       contact.History,
		OptOutAt:      contact.OptOutAt,
		LastContactAt: contact.LastContactAt,
		CreatedAt:     contact.CreatedAt,
		UpdatedAt:     contact.UpdatedAt,

		CampaignName:        campaign.Name,
		CampaignDescription: campaign.Description,

		TemplateID:           &templateID,
		Brand:                settings.Brand,
		Subject:              &settings.Subject,
		Tone:                 settings.Tone,
		EmailFrom:            &settings.EmailFrom,
		EmailReplyTo:         &settings.EmailReply,
		EmailFooter:          settings.EmailFooter,
		EmailInstructions:    &settings.EmailInstructions,
		WhatsAppFrom:         &settings.WhatsAppFrom,
		WhatsAppReplyTo:      &settings.WhatsAppReply,
		WhatsAppFooter:       settings.WhatsAppFooter,
		WhatsAppInstructions: &settings.WhatsAppInstructions,
	}

	return NewCCampaignMessageFullDTO(dto)
}

// calcularIdade retorna a idade baseada na data de nascimento
func calcularIdade(birthDate time.Time) int {
	idade := time.Now().Year() - birthDate.Year()

	// Ajusta a idade se o aniversário ainda não ocorreu neste ano
	if time.Now().YearDay() < birthDate.YearDay() {
		idade--
	}
	return idade
}
