// File: /internal/service/contact_import_service.go

package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// ContactImportService define a interface do serviço de importação
type ContactImportService interface {
	ProcessCSVAndSaveDB(ctx context.Context, inputCSV io.Reader, accountID uuid.UUID, config *dto.ConfigImportContactDTO) (int, int, error)
}

// contactImportService implementação do serviço de importação de contatos
type contactImportService struct {
	log          *slog.Logger
	contactRepo  db.ContactRepository
	openAIClient OpenAIService
}

// NewContactImportService cria uma nova instância do serviço de importação
func NewContactImportService(contactRepo db.ContactRepository, openAIClient OpenAIService) ContactImportService {
	return &contactImportService{
		log:          logger.GetLogger(),
		contactRepo:  contactRepo,
		openAIClient: openAIClient,
	}
}

// 🔹 Número de workers para processamento paralelo
const workerCount = 5

// ProcessCSVAndSaveDB processa um CSV e salva os contatos no banco
func (s *contactImportService) ProcessCSVAndSaveDB(ctx context.Context, inputCSV io.Reader, accountID uuid.UUID, config *dto.ConfigImportContactDTO) (int, int, error) {
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
	var wg sync.WaitGroup
	var recordsWG sync.WaitGroup
	successCount := 0
	failedCount := 0
	var mu sync.Mutex

	// 🔹 Inicia os workers para processar os registros
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.log.Debug("Worker iniciado", slog.Int("worker_id", workerID))

			for record := range recordsChan {
				s.log.Debug("Worker processando registro", slog.Int("worker_id", workerID))
				s.processRecord(ctx, record, headers, config, accountID, &successCount, &failedCount, &mu, &recordsWG)
				s.log.Debug("Worker finalizou processamento do registro", slog.Int("worker_id", workerID))
			}

			s.log.Debug("Worker finalizado", slog.Int("worker_id", workerID))
		}(i)
	}

	// 🔹 Enviar registros para processamento
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			s.log.Warn("Erro ao ler linha do CSV", "error", err)
			mu.Lock()
			failedCount++
			mu.Unlock()
			continue
		}

		recordsWG.Add(1)
		recordsChan <- record
	}

	recordsWG.Wait()
	close(recordsChan)
	wg.Wait()

	s.log.Info("Processamento do CSV concluído",
		slog.Int("success_count", successCount),
		slog.Int("failed_count", failedCount),
	)

	return successCount, failedCount, nil
}

// 🔹 Processar um registro individual do CSV
func (s *contactImportService) processRecord(ctx context.Context, record []string, headers []string, config *dto.ConfigImportContactDTO, accountID uuid.UUID, successCount *int, failedCount *int, mu *sync.Mutex, recordsWG *sync.WaitGroup) {
	logID := uuid.New().String()
	s.log.Debug("Iniciando processamento do registro", slog.String("log_id", logID))

	defer func() {
		s.log.Debug("Finalizando processamento do registro", slog.String("log_id", logID))
		recordsWG.Done()
	}()

	// 🔹 Gera o prompt para a IA
	prompt := GenerateContactPromptForAI(record, headers, config)

	// 🔹 Envia para OpenAI via OpenAIService
	contactDTO, err := s.formatRecordWithAI(ctx, prompt, logID)
	if err != nil {
		s.log.Error("Erro ao formatar registro",
			slog.String("log_id", logID),
			slog.String("error", err.Error()))
		mu.Lock()
		(*failedCount)++
		mu.Unlock()
		return
	}

	s.log.Debug("Contato formatado com sucesso", slog.String("log_id", logID))

	// Normaliza os dados
	contactDTO.Normalize()

	// 🔹 Verifica se o contato já existe
	existingContact, _ := s.contactRepo.FindByEmailOrWhatsApp(ctx, accountID, contactDTO.Email, contactDTO.WhatsApp)
	if existingContact != nil {
		s.log.Info("Contato já existe no banco, ignorando registro",
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

	_, err = s.contactRepo.Create(ctx, &contact)
	if err != nil {
		s.log.Error("Erro ao salvar contato no banco de dados",
			slog.String("log_id", logID),
			slog.String("error", err.Error()))
		mu.Lock()
		(*failedCount)++
		mu.Unlock()
		return
	}

	s.log.Info("Contato processado com sucesso",
		slog.String("log_id", logID),
		slog.String("account_id", accountID.String()),
		slog.String("name", contact.Name))

	mu.Lock()
	(*successCount)++
	mu.Unlock()
}

// 🔹 Envia o prompt para a OpenAI e recebe a resposta usando OpenAIService
func (s *contactImportService) formatRecordWithAI(ctx context.Context, prompt string, logID string) (*dto.ContactCreateDTO, error) {
	s.log.Debug("Enviando prompt para OpenAI",
		slog.String("log_id", logID),
		slog.String("prompt", prompt))

	request := ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: "Você é um assistente que retorna exclusivamente JSON puro, sem marcações de código, sem comentários e sem texto adicional. Apenas retorne um objeto JSON válido."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.7,
	}

	aiResponse, err := s.openAIClient.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar OpenAI: %w", err)
	}

	if len(aiResponse.Choices) == 0 {
		return nil, fmt.Errorf("nenhuma resposta válida da OpenAI")
	}

	// Extrai o JSON da resposta
	cleanJSON := utils.SanitizeJSONResponse(aiResponse.Choices[0].Message.Content)

	var contactDTO dto.ContactCreateDTO
	if err := json.Unmarshal([]byte(cleanJSON), &contactDTO); err != nil {
		return nil, fmt.Errorf("erro ao converter JSON para DTO: %w", err)
	}

	s.log.Info("Contato formatado com sucesso pela OpenAI",
		slog.String("log_id", logID))

	return &contactDTO, nil
}
