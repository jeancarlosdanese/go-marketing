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

// Controla o número máximo de requisições simultâneas para a OpenAI (10 requisições por segundo).
var rateLimiter = NewOpenAIRateLimiter(10) // Exemplo: 10 RPS

// Define que o processamento será feito em paralelo com 5 goroutines
const workerCount = 5

// Estruturas para Comunicação com a OpenAI

// OpenAIRequest representa o payload enviado para a OpenAI
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []ChatMsg `json:"messages"`
	Temperatura float64   `json:"temperature"`
}

// ChatMsg representa uma mensagem enviada para a OpenAI
type ChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse representa a resposta da OpenAI
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
	var buf bytes.Buffer
	tee := io.TeeReader(inputCSV, &buf) // Salva o conteúdo do CSV em um buffer para detectar o delimitador

	reader := csv.NewReader(tee)                      // Lê o CSV a partir do buffer
	reader.Comma = utils.DetectDelimiter(buf.Bytes()) // Detecta o delimitador do CSV
	reader.LazyQuotes = true                          // Permite erros menores de aspas. Ex: "campo com "aspas""
	reader.TrimLeadingSpace = true                    // Remove espaços em branco no início dos campos
	reader.FieldsPerRecord = -1                       // Permite número variável de campos por linha

	// Lê os cabeçalhos do CSV
	headers, err := reader.Read()
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao ler cabeçalhos do CSV: %w", err)
	}

	recordsChan := make(chan []string, workerCount) // Canal para enviar registros para as goroutines
	var wg sync.WaitGroup                           // WaitGroup para esperar todas as goroutines terminarem
	successCount := 0                               // Contador de registros processados com sucesso
	failedCount := 0                                // Contador de registros com falha
	var mu sync.Mutex

	// Inicia as goroutines para processar os registros
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for record := range recordsChan {
				processRecord(record, headers, config, contactRepo, accountID, &successCount, &failedCount, &mu)
			}
		}()
	}

	// Lê cada linha do CSV e envia para o canal recordsChan, que será consumido pelos workers.
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
		recordsChan <- record
	}

	close(recordsChan) // Fecha o canal para indicar que não há mais registros
	wg.Wait()          // Aguarda todos os workers terminarem antes de retornar

	return successCount, failedCount, nil
}

// processRecord processa um registro do CSV e salva no banco de dados
func processRecord(record []string, headers []string, config *dto.ConfigImportContactDTO, contactRepo db.ContactRepository, accountID uuid.UUID, successCount *int, failedCount *int, mu *sync.Mutex) {
	logID := uuid.New().String()

	// Gera o prompt para a AI
	prompt := GeneratePromptForAI(record, headers, config)

	// Formata o registro com a AI
	contactDTO, err := AskAIToFormatRecord(prompt, logID)
	if err != nil {
		log.Printf("[%s] Erro ao formatar registro: %v", logID, err)
		mu.Lock()
		(*failedCount)++
		mu.Unlock()
		return
	}

	contactDTO.Normalize()

	// 🔹 Agora que temos dados mais limpos, verificamos se o contato já existe no banco
	existingContact, _ := contactRepo.FindByEmailOrWhatsApp(accountID, contactDTO.Email, contactDTO.WhatsApp)
	if existingContact != nil {
		log.Printf("[%s] ✅ Contato já existe no banco. Ignorando registro.", logID)
		return
	}

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

	if contactDTO.BirthDate != nil {
		birthDate, err := time.Parse("2006-01-02", *contactDTO.BirthDate)
		if err == nil {
			contact.BirthDate = &birthDate
		}
	}

	if contactDTO.LastContactAt != nil {
		lastContactAt, err := time.Parse("2006-01-02", *contactDTO.LastContactAt)
		if err == nil {
			contact.LastContactAt = &lastContactAt
		}
	}

	_, err = contactRepo.Create(&contact)
	if err != nil {
		log.Printf("[%s] Erro ao salvar contato no DB: %v", logID, err)
		mu.Lock()
		(*failedCount)++
		mu.Unlock()
		return
	}

	mu.Lock()
	(*successCount)++
	mu.Unlock()
}

func AskAIToFormatRecord(prompt string, logID string) (*dto.ContactCreateDTO, error) {
	rateLimiter.Wait()
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY não configurada")
	}

	log.Printf("📤 [%s] Enviando para OpenAI: %s", logID, prompt)

	requestData := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMsg{
			{Role: "system", Content: "Você é um assistente que retorna apenas JSON puro."},
			{Role: "user", Content: prompt},
		},
		Temperatura: 0.7,
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

	var aiResponse OpenAIResponse
	err = json.NewDecoder(resp.Body).Decode(&aiResponse)
	if err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta da OpenAI: %w", err)
	}

	if len(aiResponse.Choices) == 0 {
		return nil, fmt.Errorf("nenhuma resposta válida da OpenAI")
	}

	rawJSON := aiResponse.Choices[0].Message.Content

	if !utils.IsValidJSON(rawJSON) {
		return nil, fmt.Errorf("JSON inválido da OpenAI")
	}

	var contactDTO dto.ContactCreateDTO
	if err := json.Unmarshal([]byte(rawJSON), &contactDTO); err != nil {
		return nil, fmt.Errorf("erro ao converter JSON para DTO: %w", err)
	}

	log.Printf("✅ [%s] Contato processado com sucesso", logID)

	return &contactDTO, nil
}
