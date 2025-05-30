// internal/dto/webhook_baileys_dto.go

package dto

type WebhookBaileysPayload struct {
	SessionID   string `json:"sessionId"`   // ID da sessão (ex: vendas_hyberica)
	From        string `json:"from"`        // JID completo (ex: 554999661111@s.whatsapp.net)
	Phone       string `json:"phone"`       // Somente o número (ex: 554999661111)
	Message     string `json:"message"`     // Texto da mensagem (se houver)
	Timestamp   int64  `json:"timestamp"`   // Timestamp Unix (ex: 1748440276)
	Datetime    string `json:"datetime"`    // ISO8601 UTC (ex: 2025-05-28T14:12:56Z)
	Type        string `json:"type"`        // Tipo da mensagem (text, image, etc)
	MessageID   string `json:"messageId"`   // ID único da mensagem
	PushName    string `json:"pushName"`    // Nome visível no WhatsApp
	FromMe      bool   `json:"fromMe"`      // true se a mensagem foi enviada por esta sessão
	IsGroup     bool   `json:"isGroup"`     // true se veio de um grupo
	Participant string `json:"participant"` // se for grupo, mostra quem enviou
}
