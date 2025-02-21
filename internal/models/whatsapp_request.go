// File: /internal/models/whatsapp_request.go

package models

// WhatsAppRequest representa uma mensagem a ser enviada via Evolution API
type WhatsAppRequest struct {
	To         string            `json:"to"`          // Número de telefone do destinatário
	TemplateID string            `json:"template_id"` // ID do template de mensagem
	Variables  map[string]string `json:"variables"`   // Variáveis para o template da mensagem
}
