// File: /internal/service/email_message_service.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type EmailMessageService interface {
	GenerateEmailContent(name, campaignName, brand, tone string) (string, error)
}

type emailMessageService struct {
	log    *slog.Logger
	client *openai.Client
}

func NewEmailMessageService() EmailMessageService {
	log := logger.GetLogger()

	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")),
	)
	return &emailMessageService{log: log, client: client}
}

func (s *emailMessageService) GenerateEmailContent(name, campaignName, brand, tone string) (string, error) {
	// System Prompt para orientar a IA
	systemPrompt := `
Você é um assistente de marketing especializado na criação de mensagens personalizadas para campanhas de e-mail.
Suas respostas devem seguir um tom adequado e engajar o destinatário para aumentar a conversão.

Diretrizes:
1. Gere um texto persuasivo e amigável, mantendo uma abordagem profissional.
2. A mensagem deve ter uma saudação, um corpo envolvente e um fechamento chamando para ação.
3. Use HTML para estruturar o conteúdo, com parágrafos bem formatados (<p>) e negrito para destaques (<strong>).
4. O tom da mensagem pode variar de acordo com o estilo da campanha (exemplo: "casual", "profissional", "promocional").
5. Não utilize emojis ou palavras exageradas como "imperdível" ou "aproveite agora" de maneira forçada.
	`

	// Usuário prompt com informações do e-mail
	userPrompt := fmt.Sprintf(`
Gere uma mensagem de e-mail com base nos seguintes detalhes:
- Nome do destinatário: %s
- Nome da campanha: %s
- Marca/empresa: %s
- Tom da mensagem: %s

Retorne um JSON puro e sanitizado:
{
    "subject": "Assunto do e-mail",
    "body": "Conteúdo formatado em HTML"
}`, name, campaignName, brand, tone)

	// Chamada à API da OpenAI
	resp, err := s.client.Chat.Completions.New(context.TODO(), openai.ChatCompletionNewParams{
		Model: openai.F(openai.ChatModelGPT4oMini),
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		}),
		MaxTokens:   openai.Int(500),
		Temperature: openai.Float(0.7),
	})
	if err != nil {
		return "", fmt.Errorf("erro ao gerar conteúdo do e-mail: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("a OpenAI não gerou uma resposta válida: %v", resp)
	}

	// Parse da resposta JSON
	var result struct {
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}
	if err := json.Unmarshal([]byte(resp.Choices[0].Message.Content), &result); err != nil {
		return "", fmt.Errorf("erro ao interpretar a resposta da OpenAI: %w", err)
	}

	return result.Body, nil
}
