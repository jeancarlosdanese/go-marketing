// internal/dto/chat_message_dto.go

package dto

type ChatMessageCreateDTO struct {
	Actor   string `json:"actor"` // "cliente", "atendente", "ai"
	Type    string `json:"type"`  // "texto", "audio", "imagem", etc.
	Content string `json:"content,omitempty"`
	FileURL string `json:"file_url,omitempty"`
}

type SendMessageDTO struct {
	PhoneNumber string `json:"phone_number"` // formato internacional
	Message     string `json:"message"`
}
