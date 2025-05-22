// internal/dto/chat_contact_dto.go

package dto

import "time"

type ChatContactResumoDTO struct {
	ID         string    `json:"id"`
	ContactID  string    `json:"contact_id"`
	Nome       string    `json:"nome"`
	WhatsApp   string    `json:"whatsapp"`
	Status     string    `json:"status"`
	Atualizado time.Time `json:"atualizado_em"`
}
