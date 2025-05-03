// File: /internal/service/campaign_processor_service.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// CampaignProcessorService define as operações para enviar campanhas para as filas SQS
type CampaignProcessorService interface {
	ProcessCampaign(ctx context.Context, campaign *models.Campaign, audience []dto.CampaignMessageDTO) error
	GenerateCampaignContent(ctx context.Context, data dto.CampaignMessageFullDTO) (*dto.CampaignContentResult, string, error)
}

type campaignProcessorService struct {
	log          *slog.Logger
	sqsService   SQSService
	openai       OpenAIService
	model        string
	audienceRepo db.CampaignAudienceRepository
}

// NewCampaignProcessorService cria um novo serviço para processar campanhas
func NewCampaignProcessorService(sqsService SQSService, openai OpenAIService, audienceRepo db.CampaignAudienceRepository) CampaignProcessorService {
	return &campaignProcessorService{
		log:          logger.GetLogger(),
		sqsService:   sqsService,
		audienceRepo: audienceRepo,
		openai:       openai,
		model:        "gpt-4o-mini", // você pode tornar isso configurável via .env se preferir
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

// 🧠 Gerar conteúdo da campanha com IA
func (s *campaignProcessorService) GenerateCampaignContent(ctx context.Context, data dto.CampaignMessageFullDTO) (*dto.CampaignContentResult, string, error) {
	prompt := buildPromptFromCampaignData(data)

	response, err := s.openai.CreateChatCompletion(ctx, ChatCompletionRequest{
		Model:       s.model,
		Temperature: 0.7,
		Messages: []ChatMessage{
			{Role: "system", Content: "Você é um especialista em marketing e personalização de mensagens para campanhas."},
			{Role: "user", Content: prompt},
		},
	})
	if err != nil {
		return nil, prompt, fmt.Errorf("erro na IA: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, prompt, fmt.Errorf("nenhuma resposta gerada pela IA")
	}

	// Extrai o JSON da resposta
	cleanJSON := utils.SanitizeJSONResponse(response.Choices[0].Message.Content)

	var result dto.CampaignContentResult
	err = json.Unmarshal([]byte(cleanJSON), &result)
	if err != nil {
		return nil, prompt, fmt.Errorf("erro ao decodificar JSON da IA: %w", err)
	}

	return &result, prompt, nil
}

// buildPromptFromCampaignData constrói o prompt a partir do DTO completo
func buildPromptFromCampaignData(data dto.CampaignMessageFullDTO) string {
	cidade := "sua região"
	if data.Cidade != nil {
		cidade = *data.Cidade
	}

	tone := "neutro"
	if data.Tone != nil {
		tone = *data.Tone
	}

	// ✅ Serializa corretamente as tags como JSON formatado
	tagsJSON := "{}"
	if len(data.Tags) > 0 {
		if b, err := json.MarshalIndent(data.Tags, "", "  "); err == nil {
			tagsJSON = string(b)
		}
	}

	return fmt.Sprintf(`Gere um JSON com os seguintes campos: saudacao, corpo, finalizacao e assinatura.
Esses campos serão usados para preencher um template de mensagem.

Contexto da campanha:
- Nome da campanha: %s
- Marca: %s
- Descrição: %s
- Tom de voz: %s
- Assunto (se email): %s

Informações do contato:
- Nome: %s
- Cidade: %s
- Estado: %s
- Gênero: %s
- Idade: %d
- Histórico: %s
- Tags:
%s

Use uma linguagem envolvente, adequada ao canal (%s) e personalizada ao contato. Responda apenas com o JSON.`,
		data.CampaignName,
		data.Brand,
		valueOrEmpty(data.CampaignDescription),
		tone,
		valueOrEmpty(data.Subject),
		data.Name,
		cidade,
		valueOrEmpty(data.Estado),
		valueOrEmpty(data.Gender),
		valueOrZero(data.Idade),
		valueOrEmpty(data.History),
		tagsJSON,
		data.Type,
	)
}

func valueOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func valueOrZero(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}
