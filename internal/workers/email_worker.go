// File: /internal/workers/email_worker.go

package workers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// EmailWorker define as operações para processar as filas de e-mails
type EmailWorker interface {
	Start(ctx context.Context)
}

// emailWorker processa mensagens de e-mail do SQS
type emailWorker struct {
	log                  *slog.Logger
	sqsService           service.SQSService
	emailService         service.EmailService
	audienceRepo         db.CampaignAudienceRepository
	contactRepo          db.ContactRepository
	campaignRepo         db.CampaignRepository
	accountRepo          db.AccountRepository
	accountSettingsRepo  db.AccountSettingsRepository
	campaignSettingsRepo db.CampaignSettingsRepository
	openAIClient         service.OpenAIService
}

// NewEmailWorker cria um novo Worker de E-mails
func NewEmailWorker(
	sqsService service.SQSService,
	emailService service.EmailService,
	audienceRepo db.CampaignAudienceRepository,
	contactRepo db.ContactRepository,
	campaignRepo db.CampaignRepository,
	accountRepo db.AccountRepository,
	accountSettingsRepo db.AccountSettingsRepository,
	campaignSettingsRepo db.CampaignSettingsRepository,
	openAIClient service.OpenAIService,
) EmailWorker {
	return &emailWorker{
		log:                  logger.GetLogger(),
		sqsService:           sqsService,
		emailService:         emailService,
		audienceRepo:         audienceRepo,
		contactRepo:          contactRepo,
		campaignRepo:         campaignRepo,
		accountRepo:          accountRepo,
		accountSettingsRepo:  accountSettingsRepo,
		campaignSettingsRepo: campaignSettingsRepo,
		openAIClient:         openAIClient,
	}
}

// Start inicia o consumo da fila de e-mails
func (w *emailWorker) Start(ctx context.Context) {
	w.log.Info("📨 EmailWorker iniciado 🚀")
	go func() {
		err := w.sqsService.ReceiveMessages(ctx, "email", func(msg dto.CampaignMessageDTO) error {
			go func() {
				err := w.processEmailMessage(ctx, msg)
				if err != nil {
					w.log.Error("❌ Erro ao processar mensagem", "error", err)
				}
			}()
			return nil
		})
		if err != nil {
			w.log.Error("❌ Erro ao iniciar processamento de mensagens", "error", err)
		}
	}()
}

// processEmailMessage processa mensagens da fila de e-mail
func (w *emailWorker) processEmailMessage(ctx context.Context, campaignMessage dto.CampaignMessageDTO) error {
	w.log.Info("📦 Processando campaignMessage", "campaign_id", campaignMessage.CampaignID, "contact_id", campaignMessage.ContactID)

	// 🔍 Buscar conta
	account, err := w.accountRepo.GetByID(ctx, campaignMessage.AccountID)
	if err != nil || account == nil {
		w.log.Error("❌ Conta não encontrada", "account_id", campaignMessage.AccountID, "error", err)
		return fmt.Errorf("conta não encontrada (account_id: %s)", campaignMessage.AccountID)
	}

	// 🔍 Buscar configurações da conta
	accountSettings, err := w.accountSettingsRepo.GetByAccountID(ctx, campaignMessage.AccountID)
	if err != nil || accountSettings == nil {
		w.log.Error("❌ Configurações da conta não encontradas", "account_id", campaignMessage.AccountID, "error", err)
		return fmt.Errorf("configurações da conta não encontradas (account_id: %s)", campaignMessage.AccountID)
	}

	// 🔍 Buscar campanha
	campaign, err := w.campaignRepo.GetByID(ctx, campaignMessage.CampaignID)
	if err != nil || campaign == nil {
		w.log.Error("❌ Campanha não encontrada", "campaign_id", campaignMessage.CampaignID, "error", err)
		return fmt.Errorf("campanha não encontrada (campaign_id: %s)", campaignMessage.CampaignID)
	}

	// 🔍 Buscar configurações da campanha
	campaignSettings, err := w.campaignSettingsRepo.GetSettingsByCampaignID(ctx, campaignMessage.CampaignID)
	if err != nil || campaignSettings == nil {
		w.log.Warn("⚠️ Configurações da campanha não encontradas, buscando última configuração usada pela conta...", "campaign_id", campaignMessage.CampaignID)

		// 🔄 Fallback: tentar buscar a última configuração usada pela conta
		campaignSettings, err = w.campaignSettingsRepo.GetLastSettings(ctx, campaignMessage.AccountID)
		if err != nil || campaignSettings == nil {
			w.log.Error("❌ Nenhuma configuração encontrada para esta campanha ou conta", "campaign_id", campaignMessage.CampaignID)
			return fmt.Errorf("nenhuma configuração encontrada para a campanha ou conta (campaign_id: %s)", campaignMessage.CampaignID)
		}
	}

	// 🔍 Buscar detalhes do contato no banco
	contact, err := w.contactRepo.GetByID(ctx, campaignMessage.ContactID)
	if err != nil || contact == nil {
		w.log.Error("❌ Contato não encontrado", "contact_id", campaignMessage.ContactID, "error", err)
		return fmt.Errorf("contato não encontrado (contact_id: %s)", campaignMessage.ContactID)
	}

	// 🔍 Validar se o contato possui e-mail
	if contact.Email == nil || *contact.Email == "" {
		w.log.Error("❌ Contato não possui e-mail válido", "contact_id", campaignMessage.ContactID)
		return fmt.Errorf("contato %s não possui e-mail válido", campaignMessage.ContactID)
	}

	// 🔹 Criar conteúdo do e-mail usando AI
	emailData, err := w.emailService.CreateEmailWithAI(ctx, *contact, *campaign, *campaignSettings)
	if err != nil || emailData == nil {
		w.log.Error("❌ Erro ao criar e-mail com AI", "contact_id", campaignMessage.ContactID, "error", err)
		return fmt.Errorf("falha ao gerar emailData (contact_id: %s)", campaignMessage.ContactID)
	}

	w.log.Info("📨 Preparando e-mail para envio", "to", *contact.Email)

	// 🚀 Enviar e-mail
	sesEmailOutput, err := w.emailService.SendEmail(*account, *accountSettings, *campaign, *campaignSettings, *contact, *emailData)
	if err != nil {
		w.log.Error("Erro ao enviar email", "error", err)
		return err
	}

	// ✅ Atualizar status para "enviado"
	w.audienceRepo.UpdateStatus(ctx, contact.ID, "enviado", *sesEmailOutput.MessageId, nil)

	w.log.Info("✅ E-mail enviado com sucesso!", "to", *sesEmailOutput.MessageId)
	return nil
}
