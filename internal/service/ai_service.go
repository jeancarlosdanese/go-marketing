// File: /internal/service/ai_service.go

package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

var rateLimiter = NewOpenAIRateLimiter(10) // Exemplo: 10 RPS

// OpenAIRequest representa o payload enviado para a OpenAI
type OpenAIRequest struct {
	Model    string    `json:"model"`
	Messages []ChatMsg `json:"messages"`
}

type ChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ProcessCSVAndSaveDB processa cada linha do CSV e salva no banco de dados
func ProcessCSVAndSaveDB(inputCSV io.Reader, contactRepo db.ContactRepository, accountID uuid.UUID, config *dto.ConfigImportContactDTO) (int, int, error) {
	// 🔍 Criar um buffer para garantir que podemos reler o CSV
	var buf bytes.Buffer
	tee := io.TeeReader(inputCSV, &buf)

	// 🔍 Abrir o CSV original com configuração mais segura
	reader := csv.NewReader(tee)
	reader.Comma = utils.DetectDelimiter(buf.Bytes()) // Detecta delimitador dinamicamente
	reader.LazyQuotes = true                          // Permite aspas inconsistentes
	reader.TrimLeadingSpace = true                    // Remove espaços extras antes de campos
	reader.FieldsPerRecord = -1                       // Permite número variável de colunas por linha

	// 🔍 Ler cabeçalhos
	headers, err := reader.Read()
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao ler cabeçalhos do CSV: %w", err)
	}

	// 🔄 Processar cada linha de forma concorrente
	var wg sync.WaitGroup
	sem := make(chan struct{}, 5) // Limita 5 requisições simultâneas
	successCount := 0
	failedCount := 0
	var mu sync.Mutex

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Erro ao ler linha do CSV:", err)
			failedCount++
			continue
		}

		wg.Add(1)
		sem <- struct{}{} // Bloqueia se houver 5 requisições ativas

		go func(record []string) {
			defer wg.Done()
			defer func() { <-sem }() // Libera o slot quando a requisição termina

			// 🔹 Gerar prompt para a IA usando a configuração
			prompt := GeneratePromptForAI(record, headers, config)

			// 🔹 Enviar para a OpenAI
			contactDTO, err := AskAIToFormatRecord(prompt)
			if err != nil {
				fmt.Println("Erro ao formatar registro:", err)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			contactDTO.Normalize() // Normaliza campos do DTO

			// 🔹 Criar modelo de contato e salvar no banco
			contact := models.Contact{
				AccountID: accountID,
				Name:      contactDTO.Name,
				Email:     contactDTO.Email,
				WhatsApp:  contactDTO.WhatsApp,
				Gender:    contactDTO.Gender,
				Bairro:    contactDTO.Bairro,
				Cidade:    contactDTO.Cidade,
				Estado:    contactDTO.Estado,
				Tags:      contactDTO.Tags,
				History:   contactDTO.History,
			}

			// 🔹 Converter data de nascimento
			if contactDTO.BirthDate != nil {
				birthDate, err := time.Parse("2006-01-02", *contactDTO.BirthDate)
				if err != nil {
					fmt.Println("Erro ao converter data de nascimento:", err)
					mu.Lock()
					failedCount++
					mu.Unlock()
					return
				}
				contact.BirthDate = &birthDate
			}

			// 🔹 Salvar no banco
			_, err = contactRepo.Create(&contact)
			if err != nil {
				fmt.Println("Erro ao salvar contato no DB:", err)
				mu.Lock()
				failedCount++
				mu.Unlock()
				return
			}

			mu.Lock()
			successCount++
			mu.Unlock()
		}(record)
	}

	wg.Wait()
	return successCount, failedCount, nil
}

// 🔹 Envia um prompt para a OpenAI e retorna um DTO formatado
func AskAIToFormatRecord(prompt string) (*dto.ContactCreateDTO, error) {
	rateLimiter.Wait() // Garante que não ultrapassamos o limite

	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY não configurada")
	}

	// 🔹 Criar requisição para OpenAI
	requestData := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMsg{
			{Role: "system", Content: "Você é um assistente que retorna apenas JSON puro, sem formatação extra."},
			{Role: "user", Content: prompt},
		},
	}

	reqBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar JSON: %w", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+openaiAPIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 🔹 Decodificando a resposta usando um struct intermediário
	var aiResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&aiResponse)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta da OpenAI: %w", err)
	}

	// 🔹 Verificar se há uma resposta válida
	if len(aiResponse.Choices) == 0 {
		return nil, fmt.Errorf("nenhuma escolha retornada pela OpenAI")
	}

	// 🔹 Extrair o conteúdo da resposta
	rawJSON := aiResponse.Choices[0].Message.Content

	// 🔹 Decodificar JSON para ContactCreateDTO
	var contactDTO dto.ContactCreateDTO
	if err := json.Unmarshal([]byte(rawJSON), &contactDTO); err != nil {
		return nil, fmt.Errorf("erro ao converter JSON para DTO: %w", err)
	}

	log.Printf("✅ Contato processado: %+v", contactDTO)

	return &contactDTO, nil
}
