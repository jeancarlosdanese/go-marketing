// internal/dto/webhook_dto.go

package dto

type WebhookPayload struct {
	Event       string      `json:"event"`
	Instance    string      `json:"instance"`
	Data        WebhookData `json:"data"`
	Sender      string      `json:"sender"`
	DateTime    string      `json:"date_time"`
	Destination string      `json:"destination"`
}

type WebhookData struct {
	Key struct {
		RemoteJID string `json:"remoteJid"`
		FromMe    bool   `json:"fromMe"`
		ID        string `json:"id"`
	} `json:"key"`

	PushName string `json:"pushName"`

	Message struct {
		Conversation string `json:"conversation"`
	} `json:"message"`

	MessageType      string `json:"messageType"`
	MessageTimestamp int64  `json:"messageTimestamp"`
	InstanceID       string `json:"instanceId"`
	Source           string `json:"source"`
}
