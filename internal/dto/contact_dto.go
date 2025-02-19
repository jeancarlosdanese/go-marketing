// File: /internal/dto/contact_dto.go

package dto

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// ContactCreateDTO define os dados para criação de um contato
type ContactCreateDTO struct {
	AccountID     uuid.UUID          `json:"account_id"` // ID da conta proprietária do contato
	Name          string             `json:"name"`
	Email         *string            `json:"email,omitempty"`
	WhatsApp      *string            `json:"whatsapp,omitempty"`
	Gender        *string            `json:"gender,omitempty"`
	BirthDate     *string            `json:"birth_date,omitempty"`
	Bairro        *string            `json:"bairro,omitempty"`
	Cidade        *string            `json:"cidade,omitempty"`
	Estado        *string            `json:"estado,omitempty"`
	Tags          models.ContactTags `json:"tags,omitempty"`
	History       *string            `json:"history,omitempty"`
	LastContactAt *string            `json:"last_contact_at,omitempty"`
}

// Validate valida os dados do ContactCreateDTO
func (c *ContactCreateDTO) Validate() error {
	if c.Name == "" {
		return errors.New("o nome é obrigatório")
	}

	if c.Email == nil && c.WhatsApp == nil {
		return errors.New("é necessário fornecer um e-mail ou WhatsApp")
	}

	if c.Email != nil {
		if err := utils.ValidateEmail(*c.Email); err != nil {
			return errors.New("e-mail inválido: " + err.Error())
		}
	}

	if c.WhatsApp != nil {
		if err := utils.ValidateWhatsApp(*c.WhatsApp); err != nil {
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

	if c.LastContactAt != nil {
		if err := utils.ValidateDate(*c.LastContactAt); err != nil {
			return err
		}
	}

	return nil
}

// Normalize normaliza os dados do ContactCreateDTO
func (c *ContactCreateDTO) Normalize() {
	// 🔹 Normaliza Nome (garante que não seja nil antes de capitalizar)
	if c.Name != "" {
		c.Name = *utils.Capitalize(&c.Name)
	}

	// 🔹 Normaliza Email (verifica nil antes de normalizar)
	if c.Email != nil {
		c.Email = utils.SanitizeEmail(c.Email)
	}

	// 🔹 Normaliza WhatsApp (verifica nil antes de normalizar)
	if c.WhatsApp != nil {
		c.WhatsApp = utils.SanitizeWhatsApp(c.WhatsApp)
	}

	// 🔹 Normaliza Gênero (verifica nil antes de normalizar)
	if c.Gender != nil {
		c.Gender = utils.NormalizeGender(c.Gender)
	}

	// 🔹 Normaliza Data de Nascimento, Bairro, Cidade e Estado
	if c.BirthDate != nil {
		c.BirthDate = utils.NormalizeBirthDate(c.BirthDate)
	}
	if c.Bairro != nil {
		c.Bairro = utils.Capitalize(c.Bairro)
	}
	if c.Cidade != nil {
		c.Cidade = utils.Capitalize(c.Cidade)
	}
	if c.Estado != nil {
		c.Estado = utils.NilIfEmpty(c.Estado)
	}

	// 🔹 Normaliza Histórico (verifica nil antes de processar)
	if c.History != nil {
		c.History = utils.NilIfEmpty(c.History)
	}

	// 🔹 Normaliza Último Contato (verifica nil antes de processar
	if c.LastContactAt != nil {
		c.LastContactAt = utils.NormalizeBirthDate(c.LastContactAt)
	}
}

// ContactUpdateDTO define os dados permitidos para atualização de um contato
type ContactUpdateDTO struct {
	Name          *string             `json:"name,omitempty"`
	Email         *string             `json:"email,omitempty"`
	WhatsApp      *string             `json:"whatsapp,omitempty"`
	Gender        *string             `json:"gender,omitempty"`
	BirthDate     *string             `json:"birth_date,omitempty"`
	Bairro        *string             `json:"bairro,omitempty"`
	Cidade        *string             `json:"cidade,omitempty"`
	Estado        *string             `json:"estado,omitempty"`
	Tags          *models.ContactTags `json:"tags,omitempty"`
	History       *string             `json:"history,omitempty"`
	OptOut        *bool               `json:"opt_out,omitempty"` // Se true, marca a data de opt-out
	LastContactAt *string             `json:"last_contact_at,omitempty"`
}

// ContactResponseDTO estrutura a resposta para um contato
type ContactResponseDTO struct {
	ID            string             `json:"id"`
	AccountID     string             `json:"account_id"`
	Name          string             `json:"name"`
	Email         *string            `json:"email,omitempty"`
	WhatsApp      *string            `json:"whatsapp,omitempty"`
	Gender        *string            `json:"gender,omitempty"`
	BirthDate     *string            `json:"birth_date,omitempty"`
	Bairro        *string            `json:"bairro,omitempty"`
	Cidade        *string            `json:"cidade,omitempty"`
	Estado        *string            `json:"estado,omitempty"`
	Tags          models.ContactTags `json:"tags,omitempty"`
	History       *string            `json:"history,omitempty"`
	OptOutAt      *string            `json:"opt_out_at,omitempty"`
	LastContactAt *string            `json:"last_contact_at,omitempty"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
}

// NewContactResponseDTO converte um modelo `Contact` para um DTO de resposta
func NewContactResponseDTO(contact *models.Contact) ContactResponseDTO {
	var birthDate, optOutAt, lastContactAt *string

	if contact.BirthDate != nil {
		formatted := contact.BirthDate.Format("2006-01-02")
		birthDate = &formatted
	}

	if contact.OptOutAt != nil {
		formatted := contact.OptOutAt.Format(time.RFC3339)
		optOutAt = &formatted
	}

	if contact.LastContactAt != nil {
		formatted := contact.LastContactAt.Format(time.RFC3339)
		lastContactAt = &formatted
	}

	return ContactResponseDTO{
		ID:            contact.ID.String(),
		AccountID:     contact.AccountID.String(),
		Name:          contact.Name,
		Email:         contact.Email,
		WhatsApp:      contact.WhatsApp,
		Gender:        contact.Gender,
		BirthDate:     birthDate,
		Bairro:        contact.Bairro,
		Cidade:        contact.Cidade,
		Estado:        contact.Estado,
		Tags:          contact.Tags,
		History:       contact.History,
		OptOutAt:      optOutAt,
		LastContactAt: lastContactAt,
		CreatedAt:     contact.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     contact.UpdatedAt.Format(time.RFC3339),
	}
}
