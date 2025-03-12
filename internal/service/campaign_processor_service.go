// File: /internal/service/campaign_processor_service.go

package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

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

		// 🔄 Tenta enviar a mensagem até 3 vezes antes de desistir
		retries := 3
		for i := 0; i < retries; i++ {
			err = s.sqsService.SendMessage(ctx, msg.Type, string(msgJSON))
			if err == nil {
				break // ✅ Sucesso, sai do loop
			}
			s.log.Warn("Retry envio para SQS", "contact_id", msg.ContactID, "tentativa", i+1, "error", err)
			time.Sleep(time.Duration(i+1) * time.Second) // ⏳ Backoff exponencial
		}

		if err != nil {
			s.log.Error("❌ Falha após retries", "contact_id", msg.ContactID, "queue", msg.Type, "error", err)
			s.audienceRepo.UpdateStatus(ctx, msg.ID, "falha", err.Error(), nil) // Marca erro no banco
		} else {
			s.audienceRepo.UpdateStatus(ctx, msg.ID, "na_fila", "", nil) // ✅ Sucesso
		}
	}

	s.log.Info("✅ Campanha enviada para processamento", "campaign_id", campaign.ID)
	return nil
}
