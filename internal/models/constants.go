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

// 🔹 Enum para status da campanha
type CampaignStatus string

const (
	StatusPendente    CampaignStatus = "pendente"    // Criada, aguardando ativação
	StatusProcessando CampaignStatus = "processando" // Enfileirando mensagens no SQS
	StatusEnviando    CampaignStatus = "enviando"    // Mensagens sendo enviadas
	StatusConcluida   CampaignStatus = "concluida"   // Campanha finalizada
	StatusCancelada   CampaignStatus = "cancelada"   // Cancelada pelo usuário
)

// Lista de status permitidos
var AllowedCampaignStatus = []CampaignStatus{StatusPendente, StatusProcessando, StatusEnviando, StatusConcluida, StatusCancelada}
