// File: /internal/service/worker_service.go

package service

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// WorkerService gerencia o consumo das filas do SQS e interage com os serviços de envio
type WorkerService struct {
	log             *slog.Logger
	sqsService      *SQSService
	emailService    *EmailService
	whatsappService *WhatsAppService
	audienceRepo    db.CampaignAudienceRepository
}

// NewWorkerService inicializa o serviço de workers
func NewWorkerService(sqsService *SQSService, emailService *EmailService, whatsappService *WhatsAppService, audienceRepo db.CampaignAudienceRepository) *WorkerService {
	return &WorkerService{
		log:             logger.GetLogger(),
		sqsService:      sqsService,
		emailService:    emailService,
		whatsappService: whatsappService,
		audienceRepo:    audienceRepo,
	}
}

// Start inicia os workers para processar as filas do SQS
func (w *WorkerService) Start() {
	w.log.Info("Iniciando Workers para processamento de filas...")

	go func() {
		w.log.Info("Worker de E-mail iniciado 🚀")
		w.sqsService.ReceiveMessages("email", w.processEmailMessage)
	}()

	go func() {
		w.log.Info("Worker de WhatsApp iniciado 🚀")
		w.sqsService.ReceiveMessages("whatsapp", w.processWhatsAppMessage)
	}()
}

// ProcessCampaign envia mensagens para a fila SQS
func (w *WorkerService) ProcessCampaign(campaign *models.Campaign, audience []dto.CampaignMessageDTO) error {
	w.log.Info("📢 Iniciando envio da campanha", "campaign_id", campaign.ID, "total_contatos", len(audience))

	for _, msg := range audience {
		// Serializar a mensagem para envio no SQS
		msgJSON, err := json.Marshal(msg)
		if err != nil {
			w.log.Error("Erro ao serializar mensagem para o SQS", "contact_id", msg.ContactID, "error", err)
			continue
		}

		// Enviar mensagem para a fila correspondente (email ou whatsapp)
		err = w.sqsService.SendMessage(msg.Type, string(msgJSON))
		if err != nil {
			w.log.Error("❌ Erro ao enviar mensagem para a fila", "contact_id", msg.ContactID, "queue", msg.Type, "error", err)
			continue
		}

		// Atualizar status do contato no banco para "na fila"
		w.audienceRepo.UpdateStatus(msg.ContactID, "na_fila", "", nil)
	}

	w.log.Info("✅ Campanha enviada para processamento", "campaign_id", campaign.ID)
	return nil
}

// processEmailMessage processa mensagens da fila de e-mail
func (w *WorkerService) processEmailMessage(campaignMessage dto.CampaignMessageDTO) error {
	w.log.Info("📦 Processando campaignMessage", "campaignMessage", campaignMessage)

	// 🔍 Verifica se o e-mail é válido antes de processar
	if campaignMessage.Email == nil || *campaignMessage.Email == "" {
		w.log.Error("❌ E-mail inválido para envio", "contact_id", campaignMessage.ContactID)
		return fmt.Errorf("e-mail inválido para o contato: %s", campaignMessage.ContactID)
	}

	// 🔥 **ADICIONAR VERIFICAÇÃO DO REMETENTE**
	fromEmail := "noreply@cetesc.com.br" // Certifique-se de que este e-mail está verificado no SES
	if fromEmail == "" {
		w.log.Error("❌ Remetente não configurado corretamente!")
		return fmt.Errorf("remetente não configurado corretamente")
	}

	w.log.Info("📨 Processando email para", "to", *campaignMessage.Email)

	// 🚀 Gera o conteúdo do email e envia
	emailBody, err := w.emailService.GenerateEmailContent(campaignMessage)
	if err != nil {
		w.log.Error("Erro ao gerar email", "contact_id", campaignMessage.ContactID, "error", err)
		return err
	}

	emailRequest := models.EmailRequest{
		From:       fromEmail,              // ✅ Adicionando remetente explícito
		To:         *campaignMessage.Email, // Agora garantimos que o valor é válido
		TemplateID: campaignMessage.CampaignID.String(),
		Subject:    "Sua campanha especial chegou!",
		Body:       emailBody,
	}

	// 🔍 **LOGAR O E-MAIL REQUEST ANTES DO ENVIO**
	w.log.Info("📧 Preparando e-mail para envio", "emailRequest", emailRequest)

	// 🚀 Enviar e-mail
	sesEmailOutput, err := w.emailService.SendEmail(emailRequest)
	if err != nil {
		w.log.Error("Erro ao enviar email", "error", err)
		return err
	}

	// ✅ Atualizar status para "enviado"
	w.audienceRepo.UpdateStatus(campaignMessage.ContactID, "enviado", *sesEmailOutput.MessageId, nil)

	w.log.Info("✅ E-mail enviado com sucesso!", "to", *campaignMessage.Email)
	return nil
}

// processWhatsAppMessage processa mensagens da fila de WhatsApp
func (w *WorkerService) processWhatsAppMessage(campaignMsg dto.CampaignMessageDTO) error {
	w.log.Info("📱 Processando envio de WhatsApp", "to", *campaignMsg.WhatsApp)

	// 🔥 Gerar conteúdo do WhatsApp com base nos dados
	whatsappVariables, err := w.whatsappService.GenerateWhatsAppContent(campaignMsg)
	if err != nil {
		w.log.Error("Erro ao gerar mensagem de WhatsApp", "error", err)
		return err
	}

	// Criar a estrutura final da mensagem do WhatsApp
	whatsappRequest := models.WhatsAppRequest{
		To:         *campaignMsg.WhatsApp,
		TemplateID: campaignMsg.CampaignID.String(),
		Variables:  whatsappVariables,
	}

	// 🚀 Enviar WhatsApp
	err = w.whatsappService.SendWhatsApp(whatsappRequest)
	if err != nil {
		w.log.Error("Erro ao enviar WhatsApp", "error", err)
		return err
	}

	// ✅ Atualizar status para "enviado"
	w.audienceRepo.UpdateStatus(campaignMsg.ContactID, "enviado", "", nil)

	w.log.Info("✅ Mensagem de WhatsApp enviada com sucesso!", "to", *campaignMsg.WhatsApp)
	return nil
}
