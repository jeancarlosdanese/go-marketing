// File: /internal/models/constants.go

package models

// 🔹 Enum para os tipos de canais disponíveis
type ChannelType string

const (
	EmailChannel    ChannelType = "email"
	WhatsappChannel ChannelType = "whatsapp"
)

// Lista de canais permitidos
var AllowedChannels = []ChannelType{EmailChannel, WhatsappChannel}

// 🔹 Enum para status de campanha
type CampaignStatus string

const (
	StatusPendente  CampaignStatus = "pendente"
	StatusAtiva     CampaignStatus = "ativa"
	StatusConcluida CampaignStatus = "concluida"
)

// Lista de status permitidos
var AllowedCampaignStatus = []CampaignStatus{StatusPendente, StatusAtiva, StatusConcluida}
