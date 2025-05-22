// internal/models/chat_message.go

package models

import "time"

type ChatMessage struct {
	ID              string     `json:"id"`
	ChatContactID   string     `json:"chat_contact_id"`
	Actor           string     `json:"actor"` // cliente, atendente, ai
	Type            string     `json:"type"`  // texto, audio, imagem, video, documento
	Content         string     `json:"content,omitempty"`
	FileURL         string     `json:"file_url,omitempty"`
	SourceProcessed bool       `json:"source_processed"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}
