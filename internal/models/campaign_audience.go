// File: /internal/models/campaign_audience.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// CampaignAudience representa um contato incluído em uma campanha específica
type CampaignAudience struct {
	ID         uuid.UUID               `json:"id"`
	CampaignID uuid.UUID               `json:"campaign_id"`
	ContactID  uuid.UUID               `json:"contact_id"`
	Type       ChannelType             `json:"type"` // "email" ou "whatsapp"
	Status     AudienceStatus          `json:"status"`
	MessageID  *string                 `json:"message_id,omitempty"`
	Feedback   *map[string]interface{} `json:"feedback_api,omitempty"` // JSONB
	CreatedAt  time.Time               `json:"created_at"`
	UpdatedAt  time.Time               `json:"updated_at"`
}

// NewCampaignAudience cria uma nova instância de CampaignAudience
func NewCampaignAudience(campaignID, contactID uuid.UUID, messageType ChannelType) *CampaignAudience {
	return &CampaignAudience{
		CampaignID: campaignID,
		ContactID:  contactID,
		Type:       messageType,
		Status:     AudiencePendente,
	}
}

// 🔹 Enum para status de campanha
type AudienceStatus string

const (
	AudiencePendente            AudienceStatus = "pendente"             // Estado inicial, não processado
	AudienceFila                AudienceStatus = "fila"                 // Na fila de processamento
	AudienceEnviado             AudienceStatus = "enviado"              // Mensagem enviada ou entregue
	AudienceFalhaRenderizacao   AudienceStatus = "falha_renderizacao"   // Erro na renderização
	AudienceRejeitado           AudienceStatus = "rejeitado"            // Rejeitado pelo SES
	AudienceDevolvido           AudienceStatus = "devolvido"            // Bounce (devolvido)
	AudienceReclamado           AudienceStatus = "reclamado"            // Complaint (reclamado)
	AudienceAtrasado            AudienceStatus = "atrasado"             // DeliveryDelay (atrasado)
	AudienceAtualizouAssinatura AudienceStatus = "atualizou_assinatura" // SubscriptionUpdate (atualização de assinatura)
)
