// File: /internal/workers/whatsapp_worker.go

package workers

import (
	"context"
	"log/slog"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

type WhatsAppWorker interface {
	Start(ctx context.Context)
}

// whatsAppWorker processa mensagens de WhatsApp do SQS
type whatsAppWorker struct {
	log                  *slog.Logger
	sqsService           service.SQSService
	whatsappService      service.WhatsAppService
	audienceRepo         db.CampaignAudienceRepository
	contactRepo          db.ContactRepository
	campaignRepo         db.CampaignRepository
	accountRepo          db.AccountRepository
	accountSettingsRepo  db.AccountSettingsRepository
	campaignSettingsRepo db.CampaignSettingsRepository
	openAIClient         service.OpenAIService
}

// NewWhatsAppWorker cria um novo Worker de WhatsApp
func NewWhatsAppWorker(
	sqsService service.SQSService,
	whatsappService service.WhatsAppService,
	audienceRepo db.CampaignAudienceRepository,
	contactRepo db.ContactRepository,
	campaignRepo db.CampaignRepository,
	accountRepo db.AccountRepository,
	accountSettingsRepo db.AccountSettingsRepository,
	campaignSettingsRepo db.CampaignSettingsRepository,
	openAIClient service.OpenAIService,
) WhatsAppWorker {
	return &whatsAppWorker{
		log:                  logger.GetLogger(),
		sqsService:           sqsService,
		whatsappService:      whatsappService,
		audienceRepo:         audienceRepo,
		contactRepo:          contactRepo,
		campaignRepo:         campaignRepo,
		accountRepo:          accountRepo,
		accountSettingsRepo:  accountSettingsRepo,
		campaignSettingsRepo: campaignSettingsRepo,
		openAIClient:         openAIClient,
	}
}

// Start inicia o consumo da fila de WhatsApp
func (w *whatsAppWorker) Start(ctx context.Context) {
	w.log.Info("WhatsAppWorker iniciado 🚀")
	w.sqsService.ReceiveMessages(ctx, "whatsapp", func(msg dto.CampaignMessageDTO) error {
		return w.processWhatsAppMessage(ctx, msg)
	})
}

// processWhatsAppMessage processa mensagens do WhatsApp
func (w *whatsAppWorker) processWhatsAppMessage(ctx context.Context, campaignMessage dto.CampaignMessageDTO) error {
	w.log.Info("📨 Processando mensagem de WhatsApp", "contact_id", campaignMessage.ContactID)

	// // 🔍 Buscar conta
	// account, err := w.accountRepo.GetByID(ctx, campaignMessage.AccountID)
	// if err != nil {
	// 	w.log.Error("❌ Erro ao buscar conta", "account_id", campaignMessage.AccountID, "error", err)
	// 	return err
	// }

	// // 🔍 Buscar configurações da conta
	// accountSettings, err := w.accountSettingsRepo.GetByAccountID(ctx, campaignMessage.AccountID)
	// if err != nil {
	// 	w.log.Error("❌ Erro ao buscar configurações da conta", "account_id", campaignMessage.AccountID, "error", err)
	// }

	// 🔍 Buscar configurações da campanha
	campaign, err := w.campaignRepo.GetByID(ctx, campaignMessage.CampaignID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar campanha", "campaign_id", campaignMessage.CampaignID, "error", err)
		return err
	}

	// // 🔍 Buscar configurações da campanha
	// campaignSettings, err := w.campaignSettingsRepo.GetSettingsByCampaignID(ctx, campaignMessage.CampaignID)
	// if err != nil {
	// 	w.log.Error("❌ Erro ao buscar configurações da campanha", "campaign_id", campaignMessage.CampaignID, "error", err)
	// 	return err
	// }

	// 🔍 Buscar detalhes do contato no banco
	contact, err := w.contactRepo.GetByID(ctx, campaignMessage.ContactID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar contato", "contact_id", campaignMessage.ContactID, "error", err)
		return err
	}

	// 🔹 Enviar mensagem via Evolution API
	channel := campaign.Channels["email"]
	whatsappRequest := models.WhatsAppRequest{
		To:         *contact.WhatsApp,
		TemplateID: channel.TemplateID.String(),
		Variables: map[string]string{
			"Nome":     contact.Name,
			"WhatsApp": *contact.WhatsApp,
		},
	}

	err = w.whatsappService.SendWhatsApp(whatsappRequest)
	if err != nil {
		w.log.Error("Erro ao enviar WhatsApp", "error", err)
		return err
	}

	w.log.Info("✅ Mensagem de WhatsApp enviada com sucesso!", "to", campaignMessage.ContactID)
	return nil
}
