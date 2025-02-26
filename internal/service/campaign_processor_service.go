// File: /internal/service/campaign_processor_service.go

package service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// CampaignProcessorService define as operações para enviar campanhas para as filas SQS
type CampaignProcessorService interface {
	ProcessCampaign(ctx context.Context, campaign *models.Campaign, audience []dto.CampaignMessageDTO) error
}

type campaignProcessorService struct {
	log          *slog.Logger
	sqsService   SQSService
	audienceRepo db.CampaignAudienceRepository
}

// NewCampaignProcessorService cria um novo serviço para processar campanhas
func NewCampaignProcessorService(sqsService SQSService, audienceRepo db.CampaignAudienceRepository) CampaignProcessorService {
	return &campaignProcessorService{
		log:          logger.GetLogger(),
		sqsService:   sqsService,
		audienceRepo: audienceRepo,
	}
}

// ProcessCampaign envia mensagens para a fila SQS
func (s *campaignProcessorService) ProcessCampaign(ctx context.Context, campaign *models.Campaign, audience []dto.CampaignMessageDTO) error {
	s.log.Info("📢 Iniciando envio da campanha", "campaign_id", campaign.ID, "total_contatos", len(audience))

	for _, msg := range audience {
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			s.log.Error("Erro ao serializar mensagem para o SQS", "contact_id", msg.ContactID, "error", err)
			continue
		}

		// Enviar mensagem para a fila correspondente (email ou whatsapp)
		err = s.sqsService.SendMessage(ctx, msg.Type, string(msgJSON))
		if err != nil {
			s.log.Error("❌ Erro ao enviar mensagem para a fila", "contact_id", msg.ContactID, "queue", msg.Type, "error", err)
			continue
		}

		// Atualizar status do contato no banco para "na fila"
		s.audienceRepo.UpdateStatus(ctx, msg.ID, "na_fila", "", nil)
	}

	s.log.Info("✅ Campanha enviada para processamento", "campaign_id", campaign.ID)
	return nil
}
