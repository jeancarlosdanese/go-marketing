// File: /internal/service/ai_service.go

package service

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// 🔹 Limite de requisições para a OpenAI (10 requisições por segundo)
var rateLimiter = NewOpenAIRateLimiter(10)

// 🔹 Número de workers para processamento paralelo
const workerCount = 5

// 🔹 Estruturas para Comunicação com a OpenAI
type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []ChatMsg `json:"messages"`
	Temperature float64   `json:"temperature"`
}

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

// 📌 Processar CSV e salvar no banco de dados
func ProcessCSVAndSaveDB(inputCSV io.Reader, contactRepo db.ContactRepository, accountID uuid.UUID, config *dto.ConfigImportContactDTO) (int, int, error) {
	log := logger.GetLogger()

	var buf bytes.Buffer
	tee := io.TeeReader(inputCSV, &buf)
	reader := csv.NewReader(tee)

	// Detectar delimitador e configurar CSV
	reader.Comma = utils.DetectDelimiter(buf.Bytes())
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	reader.FieldsPerRecord = -1

	// Ler cabeçalhos do CSV
	headers, err := reader.Read()
	if err != nil {
		return 0, 0, fmt.Errorf("erro ao ler cabeçalhos do CSV: %w", err)
	}

	recordsChan := make(chan []string, workerCount)
	var wg sync.WaitGroup        // Aguardar todos os workers terminarem
	var recordsWG sync.WaitGroup // 🔹 Novo WaitGroup para rastrear os registros processados
	successCount := 0            // Contador de registros processados com sucesso
	failedCount := 0             // Contador de registros com falha
	var mu sync.Mutex            // Mutex para proteger contadores

	// 🔹 Inicia os workers para processar os registros
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			log.Debug("Worker iniciado", slog.Int("worker_id", workerID))

			for record := range recordsChan {
				log.Debug("Worker processando registro", slog.Int("worker_id", workerID))
				processRecord(record, headers, config, contactRepo, accountID, &successCount, &failedCount, &mu, &recordsWG, log)
				log.Debug("Worker finalizou processamento do registro", slog.Int("worker_id", workerID))
			}

			log.Debug("Worker finalizado", slog.Int("worker_id", workerID))
		}(i)
	}

	// 🔹 Enviar registros para processamento
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Warn("Erro ao ler linha do CSV", "error", err)
			mu.Lock()
			failedCount++
			mu.Unlock()
			continue
		}

		recordsWG.Add(1) // 🔹 Adiciona um item ao contador de registros pendentes
		recordsChan <- record
	}

	recordsWG.Wait()   // 🔹 Aguarda todos os registros serem processados antes de fechar
	close(recordsChan) // 🔹 Agora pode fechar o canal, pois todos os registros já foram enviados

	wg.Wait() // 🔹 Espera os workers terminarem

	log.Info("Processamento do CSV concluído",
		slog.Int("success_count", successCount),
		slog.Int("failed_count", failedCount),
	)

	return successCount, failedCount, nil
}

// 📌 Processar um registro individual do CSV
func processRecord(record []string, headers []string, config *dto.ConfigImportContactDTO, contactRepo db.ContactRepository, accountID uuid.UUID, successCount *int, failedCount *int, mu *sync.Mutex, recordsWG *sync.WaitGroup, log *slog.Logger) {
	logID := uuid.New().String()
	log.Debug("Iniciando processamento do registro", slog.String("log_id", logID))

	// 🔹 Garantir que `recordsWG.Done()` sempre será chamado
	defer func() {
		log.Debug("Finalizando processamento do registro", slog.String("log_id", logID))
		recordsWG.Done() // 🔹 Agora garantimos que será chamado corretamente
	}()

	// Gera o prompt para a IA
	prompt := GeneratePromptForAI(record, headers, config)

	// Envia para OpenAI
	contactDTO, err := AskAIToFormatRecord(prompt, logID, log)
	if err != nil {
		log.Warn("Erro ao formatar registro",
			slog.String("log_id", logID),
			slog.String("error", err.Error()))
		mu.Lock()
		(*failedCount)++
		mu.Unlock()
		return
	}

	log.Debug("Contato formatado com sucesso", slog.String("log_id", logID))

	// Normaliza os dados
	contactDTO.Normalize()

	// 🔹 Verifica se o contato já existe
	existingContact, _ := contactRepo.FindByEmailOrWhatsApp(accountID, contactDTO.Email, contactDTO.WhatsApp)
	if existingContact != nil {
		log.Info("Contato já existe no banco, ignorando registro",
			slog.String("log_id", logID))
		return
	}

	// Criando o contato no banco
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
		log.Error("Erro ao salvar contato no banco de dados",
			slog.String("log_id", logID),
			slog.String("error", err.Error()))
		mu.Lock()
		(*failedCount)++
		mu.Unlock()
		return
	}

	log.Info("Contato processado com sucesso",
		slog.String("log_id", logID),
		slog.String("account_id", accountID.String()),
		slog.String("name", contact.Name))

	mu.Lock()
	(*successCount)++
	mu.Unlock()
}

// 📌 Envia o prompt para a OpenAI e recebe a resposta
func AskAIToFormatRecord(prompt string, logID string, log *slog.Logger) (*dto.ContactCreateDTO, error) {
	rateLimiter.Wait()
	openaiAPIKey := os.Getenv("OPENAI_API_KEY")
	if openaiAPIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY não configurada")
	}

	log.Debug("Enviando prompt para OpenAI",
		slog.String("log_id", logID),
		slog.String("prompt", prompt))

	requestData := OpenAIRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMsg{
			{Role: "system", Content: "Você é um assistente que retorna exclusivamente JSON puro, sem marcações de código, sem comentários e sem texto adicional. Apenas retorne um objeto JSON válido."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
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

	// Extrai o JSON da resposta
	cleanJSON := utils.SanitizeJSONResponse(aiResponse.Choices[0].Message.Content)

	if !utils.IsValidJSON(cleanJSON) {
		return nil, fmt.Errorf("JSON inválido da OpenAI")
	}

	var contactDTO dto.ContactCreateDTO
	if err := json.Unmarshal([]byte(cleanJSON), &contactDTO); err != nil {
		return nil, fmt.Errorf("erro ao converter JSON para DTO: %w", err)
	}

	log.Info("Contato formatado com sucesso pela OpenAI",
		slog.String("log_id", logID))

	return &contactDTO, nil
}
