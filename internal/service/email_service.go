// File: /internal/service/email_service.go

package service

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// EmailService gerencia o envio de e-mails via Amazon SES
type EmailService struct {
	log                 *slog.Logger
	ses                 *ses.Client
	accountSettingsRepo db.AccountSettingsRepository // 🔥 Adicionado para buscar remetente
}

// NewEmailService inicializa o serviço de e-mail
func NewEmailService(accountSettingsRepo db.AccountSettingsRepository) *EmailService {
	// 🔥 Carregar a configuração da AWS automaticamente
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		logger.GetLogger().Error("Erro ao carregar configuração AWS", "error", err)
		panic("Erro ao inicializar AWS SDK") // 🚨 Falha crítica
	}

	// Criar o cliente SES
	sesClient := ses.NewFromConfig(cfg)

	return &EmailService{
		log:                 logger.GetLogger(),
		ses:                 sesClient,
		accountSettingsRepo: accountSettingsRepo, // 🔥 Injetando o repositório
	}
}

// SendEmail envia um e-mail usando SES
func (s *EmailService) SendEmail(
	account models.Account,
	accountSettings models.AccountSettings,
	campaign models.Campaign,
	campaignSettings models.CampaignSettings,
	contact models.Contact,
	emailData dto.EmailData,
) (*ses.SendEmailOutput, error) {
	s.log.Debug("Enviando e-mail", "from", campaignSettings.EmailFrom, "to", contact.Email, "subject", campaignSettings.Subject)

	channel := campaign.Channels["email"]
	conteudoEmail, err := s.carregarTemplateEmail(channel.TemplateID.String(), emailData)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Erro ao renderizar template de e-mail: %w", err)
	}

	// aqui
	// Enviar e-mail
	input := &ses.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{strings.ToLower(*contact.Email)},
		},
		Message: &types.Message{
			Body: &types.Body{
				Html: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(conteudoEmail), // Usa o conteúdo renderizado do template
				},
				Text: &types.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(conteudoEmail), // Usa o conteúdo renderizado do template
				},
			},
			Subject: &types.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(campaignSettings.Subject),
			},
		},
		Source:               aws.String(campaignSettings.EmailFrom), // Remetente
		ConfigurationSetName: aws.String("SES-Bounce-Config"),        // 🔹 Vinculando ao Configuration Set
	}

	// 🚀 Enviar e-mail via SES
	sesEmailOutput, err := s.ses.SendEmail(context.TODO(), input)
	if err != nil {
		s.log.Error("Erro ao enviar e-mail", "error", err)
		return nil, err
	}

	s.log.Info("E-mail enviado com sucesso", "from", campaignSettings.EmailFrom, "to", contact.Email, "message_id", *sesEmailOutput.MessageId)
	return sesEmailOutput, nil
}

// carregarTemplateEmail usa o conteúdo embutido do template
func (s *EmailService) carregarTemplateEmail(templateID string, dados dto.EmailData) (string, error) {
	var emailTemplate string

	filePath := fmt.Sprintf("uploads/templates/email/%s.html", templateID)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("ERROR: Erro ao carregar template de e-mail: %w", err)
	}
	emailTemplate = string(content)

	tmpl, err := template.New("emailTemplate").Parse(emailTemplate)
	if err != nil {
		return "", fmt.Errorf("ERROR: Erro ao carregar template de e-mail embutido: %w", err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, dados); err != nil {
		return "", fmt.Errorf("ERROR: Erro ao renderizar template de e-mail: %w", err)
	}

	return body.String(), nil
}
