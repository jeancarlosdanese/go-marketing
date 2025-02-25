// File: /internal/service/openai_service.go

package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

type OpenAIService interface {
	CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error)
}

// 🔹 Estruturas para Comunicação com a OpenAI
type ChatMessage struct {
	Role    string `json:"role"`    // "system", "user", ou "assistant"
	Content string `json:"content"` // Conteúdo da mensagem
}

type ChatCompletionRequest struct {
	Model       string        `json:"model"`                 // Exemplo: "gpt-4"
	Messages    []ChatMessage `json:"messages"`              // Histórico de mensagens
	Temperature float64       `json:"temperature,omitempty"` // Opcional: Controle de aleatoriedade
	// Outros parâmetros opcionais podem ser adicionados conforme necessário
}

type ChatCompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int         `json:"index"`
		Message      ChatMessage `json:"message"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// openAIService gerencia as chamadas à OpenAI
type openAIService struct {
	log        *slog.Logger
	HTTPClient *http.Client
	APIKey     string
	BaseURL    string
}

// NewOpenAIService cria uma nova instância do serviço OpenAI
func NewOpenAIService() OpenAIService {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("❌ ERRO: OPENAI_API_KEY não configurada!")
	}

	return &openAIService{
		log:        logger.GetLogger(),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		APIKey:     apiKey,
		BaseURL:    "https://api.openai.com",
	}
}

func (client *openAIService) CreateChatCompletion(ctx context.Context, request ChatCompletionRequest) (*ChatCompletionResponse, error) {
	client.log.Info("📩 Enviando requisição para OpenAI", slog.String("model", request.Model))

	requestBody, err := json.Marshal(request)
	if err != nil {
		client.log.Error("❌ Erro ao serializar a requisição", slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao serializar a requisição: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", client.BaseURL+"/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		client.log.Error("❌ Erro ao criar a requisição HTTP", slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao criar a requisição HTTP: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HTTPClient.Do(req)
	if err != nil {
		client.log.Error("❌ Erro ao enviar requisição para OpenAI", slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao enviar a requisição: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		client.log.Error("❌ Resposta de erro da OpenAI",
			slog.Int("status_code", resp.StatusCode),
			slog.String("body", string(bodyBytes)),
		)
		return nil, fmt.Errorf("resposta da API com status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var chatResponse ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResponse); err != nil {
		client.log.Error("❌ Erro ao decodificar resposta da OpenAI", slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao decodificar a resposta: %w", err)
	}

	client.log.Info("✅ Resposta da OpenAI recebida com sucesso", slog.Int("tokens_utilizados", chatResponse.Usage.TotalTokens))
	return &chatResponse, nil
}
