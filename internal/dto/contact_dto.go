// File: /internal/dto/contact_dto.go

package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// ContactCreateDTO define os dados necessários para criar um contato
type ContactCreateDTO struct {
	AccountID uuid.UUID `json:"account_id"` // ID da conta proprietária do contato
	Name      string    `json:"name"`
	Email     string    `json:"email,omitempty"`
	WhatsApp  string    `json:"whatsapp,omitempty"`
	Gender    *string   `json:"gender,omitempty"`
	BirthDate *string   `json:"birth_date,omitempty"` // Formatado como "YYYY-MM-DD"
	History   *string   `json:"history,omitempty"`
}

// ContactUpdateDTO define os dados permitidos para atualizar um contato
type ContactUpdateDTO struct {
	Name      *string `json:"name,omitempty"`
	Email     *string `json:"email,omitempty"`
	WhatsApp  *string `json:"whatsapp,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
	History   *string `json:"history,omitempty"`
	OptOut    *bool   `json:"opt_out,omitempty"` // Se true, marca a data de opt-out
}

// ContactResponseDTO estrutura a resposta para um contato
type ContactResponseDTO struct {
	ID        string  `json:"id"`
	AccountID string  `json:"account_id"`
	Name      string  `json:"name"`
	Email     string  `json:"email,omitempty"`
	WhatsApp  string  `json:"whatsapp,omitempty"`
	Gender    *string `json:"gender,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
	History   *string `json:"history,omitempty"`
	OptOutAt  *string `json:"opt_out_at,omitempty"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

// NewContactResponseDTO converte um modelo `Contact` para um DTO de resposta
func NewContactResponseDTO(contact *models.Contact) ContactResponseDTO {
	var birthDate, optOutAt *string

	if contact.BirthDate != nil {
		formatted := contact.BirthDate.Format("2006-01-02")
		birthDate = &formatted
	}

	if contact.OptOutAt != nil {
		formatted := contact.OptOutAt.Format(time.RFC3339)
		optOutAt = &formatted
	}

	return ContactResponseDTO{
		ID:        contact.ID.String(),
		AccountID: contact.AccountID.String(),
		Name:      contact.Name,
		Email:     contact.Email,
		WhatsApp:  contact.WhatsApp,
		Gender:    contact.Gender,
		BirthDate: birthDate,
		History:   contact.History,
		OptOutAt:  optOutAt,
		CreatedAt: contact.CreatedAt.Format(time.RFC3339),
		UpdatedAt: contact.UpdatedAt.Format(time.RFC3339),
	}
}

// 🔍 Validações

// Validate valida os dados do ContactCreateDTO
func (c *ContactCreateDTO) Validate() error {
	if c.Name == "" {
		return errors.New("o nome é obrigatório")
	}

	if c.Email == "" && c.WhatsApp == "" {
		return errors.New("é necessário fornecer um e-mail ou WhatsApp")
	}

	if c.Email != "" {
		if err := utils.ValidateEmail(c.Email); err != nil {
			return errors.New("e-mail inválido: " + err.Error())
		}
	}

	if c.WhatsApp != "" {
		if err := utils.ValidateWhatsApp(c.WhatsApp); err != nil {
			return errors.New("número de WhatsApp inválido: " + err.Error())
		}
	}

	if c.Gender != nil {
		if err := utils.ValidateGender(*c.Gender); err != nil {
			return err
		}
	}

	if c.BirthDate != nil {
		if err := utils.ValidateDate(*c.BirthDate); err != nil {
			return err
		}
	}

	return nil
}

// Validate valida os dados do ContactUpdateDTO
func (c *ContactUpdateDTO) Validate() error {
	if c.Email != nil && utils.ValidateEmail(*c.Email) != nil {
		return errors.New("e-mail inválido")
	}

	if c.WhatsApp != nil && utils.ValidateWhatsApp(*c.WhatsApp) != nil {
		return errors.New("número de WhatsApp inválido")
	}

	if c.Gender != nil && utils.ValidateGender(*c.Gender) != nil {
		return errors.New("gênero inválido")
	}

	if c.BirthDate != nil && utils.ValidateDate(*c.BirthDate) != nil {
		return errors.New("data de nascimento inválida (use YYYY-MM-DD)")
	}

	return nil
}
