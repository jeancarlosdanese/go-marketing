// File: /internal/dto/template_dto.go

package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// TemplateCreateDTO define os dados necessários para criar um template
type TemplateCreateDTO struct {
	AccountID   uuid.UUID `json:"account_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
}

// TemplateUpdateDTO define os dados permitidos para atualizar um template
type TemplateUpdateDTO struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// TemplateResponseDTO estrutura a resposta para um template
type TemplateResponseDTO struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"account_id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// NewTemplateResponseDTO converte um modelo `Template` para um DTO de resposta
func NewTemplateResponseDTO(template *models.Template) TemplateResponseDTO {
	return TemplateResponseDTO{
		ID:          template.ID.String(),
		AccountID:   template.AccountID.String(),
		Name:        template.Name,
		Description: template.Description,
		CreatedAt:   template.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   template.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// 🔍 Validações

// Validate valida os dados do TemplateCreateDTO
func (t *TemplateCreateDTO) Validate() error {
	if t.Name == "" {
		return errors.New("o nome do template é obrigatório")
	}
	return nil
}

// Validate valida os dados do TemplateUpdateDTO
func (t *TemplateUpdateDTO) Validate() error {
	if t.Name != nil && *t.Name == "" {
		return errors.New("o nome do template não pode ser vazio")
	}
	return nil
}
