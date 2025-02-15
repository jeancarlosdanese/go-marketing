// File: /internal/models/template.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// Template representa um modelo de mensagem para campanhas
type Template struct {
	ID               uuid.UUID `json:"id"`
	AccountID        uuid.UUID `json:"account_id"`
	Name             string    `json:"name"`
	Description      *string   `json:"description,omitempty"`
	TemplateHTML     *string   `json:"template_html,omitempty"`
	TemplateWhatsApp *string   `json:"template_whatsapp,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
