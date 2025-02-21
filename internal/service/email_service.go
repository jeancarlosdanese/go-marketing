// File: /internal/service/email_service.go

package service

import (
	"context"
	"fmt"
	"log/slog"

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
func (e *EmailService) SendEmail(emailRequest models.EmailRequest) (*ses.SendEmailOutput, error) {
	// 💌 Criar a mensagem SES
	input := &ses.SendEmailInput{
		Source: &emailRequest.From, // 🔥 Agora é dinâmico
		Destination: &types.Destination{
			ToAddresses: []string{emailRequest.To},
		},
		Message: &types.Message{
			Subject: &types.Content{Data: &emailRequest.Subject},
			Body: &types.Body{
				Html: &types.Content{Data: &emailRequest.Body},
			},
		},
		ConfigurationSetName: aws.String("SES-Bounce-Config"), // 🔹 Vinculando ao Configuration Set
	}

	// 🚀 Enviar e-mail via SES
	sesEmailOutput, err := e.ses.SendEmail(context.TODO(), input)
	if err != nil {
		e.log.Error("Erro ao enviar e-mail", "error", err)
		return nil, err
	}

	e.log.Info("E-mail enviado com sucesso", "from", emailRequest.From, "to", emailRequest.To, "subject", emailRequest.Subject)
	return sesEmailOutput, nil
}

// GenerateEmailContent cria um e-mail personalizado com base nos dados do contato
func (e *EmailService) GenerateEmailContent(msg dto.CampaignMessageDTO) (string, error) {
	e.log.Info("Gerando conteúdo do e-mail", "contact_id", msg.ContactID)

	// Personalização do conteúdo
	name := msg.Name
	if name == "" {
		name = "Caro cliente"
	}

	subject := fmt.Sprintf("Olá %s, temos uma oferta especial para você!", name)
	body := fmt.Sprintf(`
		<p>Olá %s,</p>
		<p>Estamos entrando em contato para apresentar uma oferta exclusiva!</p>
		<p>Não perca essa oportunidade.</p>
		<p>Atenciosamente,<br>Equipe Marketing</p>`,
		name,
	)

	return fmt.Sprintf("Assunto: %s\n\n%s", subject, body), nil
}
