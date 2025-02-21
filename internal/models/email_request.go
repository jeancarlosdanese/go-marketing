// File: /internal/models/email_request.go

package models

import "github.com/google/uuid"

// EmailRequest representa um e-mail a ser enviado via Amazon SES
type EmailRequest struct {
	AccountID  uuid.UUID `json:"account_id"` // 🔥 Necessário para buscar o remetente correto
	To         string    `json:"to"`
	From       string    `json:"from,omitempty"` // Opcional, se não for enviado, buscamos do banco
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	TemplateID string    `json:"template_id"`
}
