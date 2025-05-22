// internal/dto/chat_dto.go

package dto

import (
	"errors"

	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type ChatCreateDTO struct {
	Department        string `json:"department"` // financeiro, comercial, suporte
	Title             string `json:"title"`
	Instructions      string `json:"instructions"`
	PhoneNumber       string `json:"phone_number"`
	EvolutionInstance string `json:"evolution_instance"`
	WebhookURL        string `json:"webhook_url"`
}

// Validate valida os dados do ContactCreateDTO
func (c *ChatCreateDTO) Validate() error {
	if c.Department == "" || len(c.Department) < 3 || len(c.Department) > 50 {
		return errors.New("o setor deve ter entre 3 e 50 caracteres")
	}

	if c.Title == "" || len(c.Title) < 3 || len(c.Title) > 150 {
		return errors.New("o título deve ter entre 3 e 150 caracteres")
	}

	if c.Instructions == "" {
		return errors.New("as instruções são obrigatórias")
	}

	if c.PhoneNumber == "" || utils.ValidateWhatsApp(c.PhoneNumber) != nil {
		return errors.New("o número de telefone deve ser um número válido")
	}

	if c.EvolutionInstance == "" || len(c.EvolutionInstance) < 3 || len(c.EvolutionInstance) > 50 {
		return errors.New("a instância de evolução deve ter entre 3 e 50 caracteres")
	}

	if c.WebhookURL == "" || len(c.WebhookURL) < 3 || len(c.WebhookURL) > 255 {
		return errors.New("a URL do webhook deve ter entre 3 e 255 caracteres")
	}

	return nil
}

// ConvertToChat converte o DTO para o modelo de Chat
func (c *ChatCreateDTO) ToModel() *models.Chat {
	// Formata o número de telefone
	phoneNumber := utils.FormatWhatsApp(c.PhoneNumber)

	return &models.Chat{
		Department:        c.Department,
		Title:             c.Title,
		Instructions:      c.Instructions,
		PhoneNumber:       phoneNumber,
		EvolutionInstance: c.EvolutionInstance,
		WebhookURL:        c.WebhookURL,
	}
}

type ChatUpdateDTO struct {
	Title             string `json:"title"`
	Instructions      string `json:"instructions"`
	PhoneNumber       string `json:"phone_number"`
	EvolutionInstance string `json:"evolution_instance"`
	WebhookURL        string `json:"webhook_url"`
}

// Validate valida os dados do ChatUpdateDTO
func (c *ChatUpdateDTO) Validate() error {
	if c.Title == "" || len(c.Title) < 3 || len(c.Title) > 150 {
		return errors.New("o título deve ter entre 3 e 150 caracteres")
	}
	if c.Instructions == "" {
		return errors.New("as instruções são obrigatórias")
	}
	if c.PhoneNumber == "" || utils.ValidateWhatsApp(c.PhoneNumber) != nil {
		return errors.New("o número de telefone deve ser válido")
	}
	if c.EvolutionInstance == "" || len(c.EvolutionInstance) < 3 || len(c.EvolutionInstance) > 50 {
		return errors.New("a instância de evolução deve ter entre 3 e 50 caracteres")
	}
	if c.WebhookURL == "" || len(c.WebhookURL) < 3 || len(c.WebhookURL) > 255 {
		return errors.New("a URL do webhook deve ter entre 3 e 255 caracteres")
	}
	return nil
}
