// internal/dto/webhook_baileys_dto.go

package dto

type WebhookBaileysPayload struct {
	SessionID string `json:"sessionId"`
	From      string `json:"from"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
}
