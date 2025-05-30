// internal/dto/chat_contact.go

package dto

// ChatContactFull representa um contato completo de chat, incluindo informações do WhatsApp
type ChatContactFull struct {
	ID                string `json:"id"`
	ChatID            string `json:"chat_id"`
	ContactID         string `json:"contact_id"`
	WhatsappContactID string `json:"whatsapp_contact_id"`
	Name              string `json:"name"`
	Phone             string `json:"phone"`
	JID               string `json:"jid"`
	IsBusiness        bool   `json:"is_business"`
	Status            string `json:"status"`     // "aberto", "fechado", "pendente"
	UpdatedAt         string `json:"updated_at"` // ISO timestamp
}
