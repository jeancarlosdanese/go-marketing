// File: /internal/service/worker_service.go

package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// WorkerService define as operações para processar as filas do SQS
type WorkerService interface {
	Start(ctx context.Context)
	ProcessCampaign(ctx context.Context, campaign *models.Campaign, audience []dto.CampaignMessageDTO) error
}

// workerService gerencia o consumo das filas do SQS e interage com os serviços de envio
type workerService struct {
	log                  *slog.Logger
	sqsService           SQSService
	emailService         *EmailService
	whatsappService      *WhatsAppService
	audienceRepo         db.CampaignAudienceRepository
	contactRepo          db.ContactRepository
	campaignRepo         db.CampaignRepository
	accountRepo          db.AccountRepository
	accountSettingsRepo  db.AccountSettingsRepository
	campaignSettingsRepo db.CampaignSettingsRepository
	openAIClient         OpenAIService
}

// NewWorkerService inicializa o serviço de workers
func NewWorkerService(
	sqsService SQSService,
	emailService *EmailService,
	whatsappService *WhatsAppService,
	audienceRepo db.CampaignAudienceRepository,
	contactRepo db.ContactRepository,
	campaignRepo db.CampaignRepository,
	accountRepo db.AccountRepository,
	accountSettingsRepo db.AccountSettingsRepository,
	campaignSettingsRepo db.CampaignSettingsRepository,
	openAIClient OpenAIService,
) WorkerService {
	return &workerService{
		log:                  logger.GetLogger(),
		sqsService:           sqsService,
		emailService:         emailService,
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

// Start inicia os workers para processar as filas do SQS
func (w *workerService) Start(ctx context.Context) {
	w.log.Info("Iniciando Workers para processamento de filas...")

	go func() {
		w.log.Info("Worker de E-mail iniciado 🚀")
		w.sqsService.ReceiveMessages(ctx, "email", func(msg dto.CampaignMessageDTO) error {
			return w.processEmailMessage(ctx, msg)
		})
	}()

	go func() {
		w.log.Info("Worker de WhatsApp iniciado 🚀")
		w.sqsService.ReceiveMessages(ctx, "whatsapp", func(msg dto.CampaignMessageDTO) error {
			return w.processWhatsAppMessage(ctx, msg)
		})
	}()
}

// ProcessCampaign envia mensagens para a fila SQS
func (w *workerService) ProcessCampaign(ctx context.Context, campaign *models.Campaign, audience []dto.CampaignMessageDTO) error {
	w.log.Info("📢 Iniciando envio da campanha", "campaign_id", campaign.ID, "total_contatos", len(audience))

	for _, msg := range audience {
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			w.log.Error("Erro ao serializar mensagem para o SQS", "contact_id", msg.ContactID, "error", err)
			continue
		}

		// Enviar mensagem para a fila correspondente (email ou whatsapp)
		err = w.sqsService.SendMessage(ctx, msg.Type, string(msgJSON))
		if err != nil {
			w.log.Error("❌ Erro ao enviar mensagem para a fila", "contact_id", msg.ContactID, "queue", msg.Type, "error", err)
			continue
		}

		// Atualizar status do contato no banco para "na fila"
		w.audienceRepo.UpdateStatus(ctx, msg.ID, "na_fila", "", nil)
	}

	w.log.Info("✅ Campanha enviada para processamento", "campaign_id", campaign.ID)
	return nil
}

// processEmailMessage processa mensagens da fila de e-mail
func (w *workerService) processEmailMessage(ctx context.Context, campaignMessage dto.CampaignMessageDTO) error {
	w.log.Info("📦 Processando campaignMessage", "campaign_id", campaignMessage.CampaignID, "contact_id", campaignMessage.ContactID)

	// 🔍 Buscar detalhes do contato no banco
	contact, err := w.contactRepo.GetByID(ctx, campaignMessage.ContactID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar contato", "contact_id", campaignMessage.ContactID, "error", err)
		return err
	}

	// 🔍 Buscar configurações da campanha
	campaign, err := w.campaignRepo.GetByID(ctx, campaignMessage.CampaignID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar campanha", "campaign_id", campaignMessage.CampaignID, "error", err)
		return err
	}

	// 🔍 Buscar configurações da campanha
	campaignSettings, err := w.campaignSettingsRepo.GetSettingsByCampaignID(ctx, campaignMessage.CampaignID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar configurações da campanha", "campaign_id", campaignMessage.CampaignID, "error", err)
		return err
	}

	// Gerar conteúdo do e-mail com AI retornando um dto.EmailData
	prompt := GenerateEmailPromptForAI(*contact, *campaign, *campaignSettings)

	emailData, err := w.createEmailWithAI(ctx, prompt)
	if err != nil {
		w.log.Error("❌ Erro ao criar e-mail com AI", "contact_id", campaignMessage.ContactID, "error", err)
	}

	// 🔍 Buscar conta
	account, err := w.accountRepo.GetByID(ctx, campaign.AccountID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar conta", "account_id", campaign.AccountID, "error", err)
		return err
	}

	// 🔍 Buscar configurações da conta
	accountSettings, err := w.accountSettingsRepo.GetByAccountID(ctx, campaign.AccountID)
	if err != nil {
		w.log.Error("❌ Erro ao buscar configurações da conta", "account_id", campaign.AccountID, "error", err)
	}

	// 🔍 Validar se o contato possui e-mail
	if contact.Email == nil {
		w.log.Error("❌ Contato não possui e-mail válido", "contact_id", campaignMessage.ContactID)
		return fmt.Errorf("contato %s não possui e-mail válido", campaignMessage.ContactID)
	}

	w.log.Info("📨 Preparando e-mail para envio", "to", contact.Email)

	// 🚀 Enviar e-mail
	sesEmailOutput, err := w.emailService.SendEmail(*account, *accountSettings, *campaign, *campaignSettings, *contact, *emailData)
	if err != nil {
		w.log.Error("Erro ao enviar email", "error", err)
		return err
	}

	// // ✅ Atualizar status para "enviado"
	// w.audienceRepo.UpdateStatus(campaignMessage.ID, "enviado", *sesEmailOutput.MessageId, nil)

	w.log.Info("✅ E-mail enviado com sucesso!", "to", sesEmailOutput.MessageId)
	return nil
}

// processWhatsAppMessage processa mensagens da fila de WhatsApp
func (w *workerService) processWhatsAppMessage(ctx context.Context, campaignMessage dto.CampaignMessageDTO) error {
	w.log.Info("📱 Processando envio de WhatsApp", "campaign_id", campaignMessage.CampaignID, "contact_id", campaignMessage.ContactID)

	// // 🔍 Buscar detalhes do contato no banco
	// contact, err := w.contactRepo.GetByID(campaignMessage.ContactID)
	// if err != nil {
	// 	w.log.Error("❌ Erro ao buscar contato", "contact_id", campaignMessage.ContactID, "error", err)
	// 	return err
	// }

	// // 🔍 Buscar configurações da campanha
	// campaign, err := w.campaignRepo.GetByID(campaignMessage.CampaignID)
	// if err != nil {
	// 	w.log.Error("❌ Erro ao buscar campanha", "campaign_id", campaignMessage.CampaignID, "error", err)
	// 	return err
	// }

	// // 🔍 Validar se o contato possui WhatsApp
	// if contact.WhatsApp == "" {
	// 	w.log.Error("❌ Contato não possui WhatsApp válido", "contact_id", campaignMessage.ContactID)
	// 	return fmt.Errorf("contato %s não possui WhatsApp válido", campaignMessage.ContactID)
	// }

	// w.log.Info("📩 Preparando mensagem de WhatsApp para envio", "to", contact.WhatsApp)

	// // 🚀 Gerar conteúdo da mensagem do WhatsApp
	// whatsappVariables, err := w.whatsappService.GenerateWhatsAppContent(contact, campaign)
	// if err != nil {
	// 	w.log.Error("Erro ao gerar mensagem de WhatsApp", "error", err)
	// 	return err
	// }

	// whatsappRequest := models.WhatsAppRequest{
	// 	To:         contact.WhatsApp,
	// 	TemplateID: campaignMessage.CampaignID.String(),
	// 	Variables:  whatsappVariables,
	// }

	// // 🚀 Enviar WhatsApp
	// err = w.whatsappService.SendWhatsApp(whatsappRequest)
	// if err != nil {
	// 	w.log.Error("Erro ao enviar WhatsApp", "error", err)
	// 	return err
	// }

	// // ✅ Atualizar status para "enviado"
	// w.audienceRepo.UpdateStatus(campaignMessage.ID, "enviado", "", nil)

	// w.log.Info("✅ Mensagem de WhatsApp enviada com sucesso!", "to", contact.WhatsApp)
	return nil
}

// 🔹 Envia o prompt para a OpenAI e recebe a resposta usando OpenAIService
func (w *workerService) createEmailWithAI(ctx context.Context, prompt string) (*dto.EmailData, error) {
	w.log.Debug("Enviando prompt para OpenAI",
		slog.String("prompt", prompt))

	request := ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: "Você é um assistente especialiado em gerar mensagens de email persuasivas, retornando exclusivamente JSON puro, sem marcações de código, sem comentários e sem texto adicional. Apenas retorne um objeto JSON válido."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
	}

	aiResponse, err := w.openAIClient.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar OpenAI: %w", err)
	}

	if len(aiResponse.Choices) == 0 {
		return nil, fmt.Errorf("nenhuma resposta válida da OpenAI")
	}

	// Extrai o JSON da resposta
	cleanJSON := utils.SanitizeJSONResponse(aiResponse.Choices[0].Message.Content)

	var emailDTO dto.EmailData
	if err := json.Unmarshal([]byte(cleanJSON), &emailDTO); err != nil {
		return nil, fmt.Errorf("erro ao converter JSON para DTO: %w", err)
	}

	w.log.Debug("✅ Email criado e formatado com sucesso pela OpenAI", "email_data", emailDTO)

	return &emailDTO, nil
}
