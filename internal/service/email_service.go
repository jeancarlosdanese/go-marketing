// File: /internal/service/email_service.go

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type EmailService interface {
	CreateEmailWithAI(ctx context.Context, contact models.Contact, campaign models.Campaign, campaignSettings models.CampaignSettings) (*dto.EmailData, error)
	SendEmail(account models.Account, accountSettings models.AccountSettings, campaign models.Campaign, campaignSettings models.CampaignSettings, contact models.Contact, emailData dto.EmailData) (*ses.SendEmailOutput, error)
}

type emailService struct {
	log    *slog.Logger
	openAI OpenAIService
}

func NewEmailService(openAIService OpenAIService) EmailService {
	log := logger.GetLogger()

	return &emailService{log: log, openAI: openAIService}
}

// 🔹 Envia o prompt para a OpenAI e recebe a resposta usando OpenAIService
func (s *emailService) CreateEmailWithAI(ctx context.Context, contact models.Contact, campaign models.Campaign, campaignSettings models.CampaignSettings) (*dto.EmailData, error) {
	s.log.Debug("Enviando prompt para OpenAI", "contact_id", contact.ID, "campaign_id", campaign.ID)

	// 🔹 Cria o prompt para a OpenAI
	prompt := s.generateEmailPromptForAI(contact, campaign, campaignSettings)

	request := ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: "Você é um assistente especialiado em gerar mensagens de email persuasivas, retornando exclusivamente JSON puro, sem marcações de código, sem comentários e sem texto adicional. Apenas retorne um objeto JSON válido."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
	}

	aiResponse, err := s.openAI.CreateChatCompletion(ctx, request)
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

	s.log.Debug("✅ Email criado e formatado com sucesso pela OpenAI", "email_data", emailDTO)

	return &emailDTO, nil
}

// SendEmail envia um e-mail usando SES
func (s *emailService) SendEmail(
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

	// Configure ses client from account settings
	awsConfig := aws.Config{
		Region: accountSettings.AWSRegion,
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			accountSettings.AWSAccessKeyID, accountSettings.AWSSecretAccessKey, ""),
		),
	}
	ses := ses.NewFromConfig(awsConfig)

	// 🚀 Enviar e-mail via SES
	sesEmailOutput, err := ses.SendEmail(context.TODO(), input)
	if err != nil {
		s.log.Error("Erro ao enviar e-mail", "error", err)
		return nil, err
	}

	s.log.Info("E-mail enviado com sucesso", "from", campaignSettings.EmailFrom, "to", contact.Email, "message_id", *sesEmailOutput.MessageId)
	return sesEmailOutput, nil
}

// GenerateEmailPromptForAI gera um prompt dinâmico para a IA criar um e-mail personalizado.
func (s *emailService) generateEmailPromptForAI(contact models.Contact, campaign models.Campaign, campaignSettings models.CampaignSettings) string {
	// 🔹 Criar um mapa com as informações do contato
	contactData := map[string]interface{}{
		"Nome":           contact.Name,
		"E-mail":         contact.Email,
		"WhatsApp":       contact.WhatsApp,
		"Gênero":         contact.Gender,
		"Histórico":      contact.History,
		"Último Contato": contact.LastContactAt,
		"Localização":    fmt.Sprintf("%s, %s - %s", utils.SafeString(contact.Bairro), utils.SafeString(contact.Cidade), utils.SafeString(contact.Estado)),
		"Interesses":     contact.Tags,
	}

	// 🔹 Criar um mapa com as informações da campanha
	campaignData := map[string]interface{}{
		"Nome":       campaign.Name,
		"Descrição":  campaign.Description,
		"Marca":      campaignSettings.Brand,
		"Assunto":    campaignSettings.Subject,
		"Tom de Voz": campaignSettings.Tone,
		"Instruções": campaignSettings.EmailInstructions,
		"Rodapé":     campaignSettings.EmailFooter,
	}

	// 🔹 Converter os mapas para JSON formatado
	contactJSON, _ := json.MarshalIndent(contactData, "", "  ")
	campaignJSON, _ := json.MarshalIndent(campaignData, "", "  ")

	// 🔹 Criar o esquema esperado para o e-mail
	emailSchema := `
	O e-mail gerado deve conter:
	1. **Saudação**: Cumprimentar o destinatário de forma personalizada.
	2. **Corpo**: Contextualizar a campanha e destacar a oferta ou mensagem principal.
	3. **Finalização**: Incentivar a ação desejada, como responder ao e-mail ou acessar um link.
	4. **Assinatura**: Finalizar com a marca da empresa e informações de contato.`

	// 🔹 Construção do prompt final
	prompt := fmt.Sprintf(`
	📩 **Geração de E-mail Personalizado**

	📌 **Detalhes do Contato**:
	%s

	📌 **Detalhes da Campanha**:
	%s

	📌 **Formato Esperado**:
	%s

	💡 **Tarefa**
	- Gere um e-mail personalizado com base nas informações acima.
	- Utilize um tom coerente com a campanha.
	- Retorne o e-mail formatado exclusivamente como um JSON válido, contendo os seguintes campos:
		{
			"saudacao": "string",
			"corpo": "string",
			"finalizacao": "string",
			"assinatura": "string"
		}
	`, string(contactJSON), string(campaignJSON), emailSchema)

	return prompt
}

// carregarTemplateEmail usa o conteúdo embutido do template
func (s *emailService) carregarTemplateEmail(templateID string, dados dto.EmailData) (string, error) {
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
