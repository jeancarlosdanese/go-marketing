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
	"strings"
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
	ProcessImport(ctx context.Context, accountID uuid.UUID, importData *models.ContactImport) error
	ProcessCSVAndSaveDB(ctx context.Context, inputCSV io.Reader, accountID uuid.UUID, config *dto.ConfigImportContactDTO) (int, int, error)
	GenerateImportConfig(ctx context.Context, headers []string, sampleRecords [][]string) (*models.ContactImportConfig, error)
}

// contactImportService implementação do serviço de importação de contatos
type contactImportService struct {
	log               *slog.Logger
	contactRepo       db.ContactRepository
	contactImportRepo db.ContactImportRepository
	openAIClient      OpenAIService
}

// NewContactImportService cria uma nova instância do serviço de importação
func NewContactImportService(contactRepo db.ContactRepository, contactImportRepo db.ContactImportRepository, openAIClient OpenAIService) ContactImportService {
	return &contactImportService{
		log:               logger.GetLogger(),
		contactRepo:       contactRepo,
		contactImportRepo: contactImportRepo,
		openAIClient:      openAIClient,
	}
}

// 🔹 Número de workers para processamento paralelo
const workerCount = 5

// ProcessImport processa uma importação salva previamente
func (s *contactImportService) ProcessImport(ctx context.Context, accountID uuid.UUID, importData *models.ContactImport) error {
	s.log.Info("Iniciando processamento da importação", slog.String("import_id", importData.ID.String()))

	// 🔹 Localiza o arquivo CSV
	file, err := utils.OpenImportContactsFile(importData.FileName)
	if err != nil {
		s.log.Error("Erro ao abrir o arquivo de importação", slog.String("error", err.Error()))
		_ = s.contactImportRepo.UpdateStatus(ctx, importData.ID, "erro")
		return err
	}
	defer file.Close()

	// 🔹 Converte config (models.ContactImportConfig) → dto.ConfigImportContactDTO
	configDTO := dto.ConvertToConfigImportContactDTO(*importData.Config)

	// 🔹 Processa e salva os contatos
	successCount, failedCount, err := s.ProcessCSVAndSaveDB(ctx, file, accountID, configDTO)
	if err != nil {
		s.log.Error("Erro ao processar CSV", slog.String("error", err.Error()))
		_ = s.contactImportRepo.UpdateStatus(ctx, importData.ID, "erro")
		return err
	}

	s.log.Info("Processamento finalizado",
		slog.Int("sucesso", successCount),
		slog.Int("falha", failedCount),
		slog.String("import_id", importData.ID.String()))

	// 🔹 Atualiza status no banco
	if successCount > 0 {
		_ = s.contactImportRepo.UpdateStatus(ctx, importData.ID, "concluido")
	} else {
		_ = s.contactImportRepo.UpdateStatus(ctx, importData.ID, "sem_dados")
	}

	return nil
}

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
		Tags:      &contactDTO.Tags,
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

func (s *contactImportService) GenerateImportConfig(ctx context.Context, headers []string, sampleRecords [][]string) (*models.ContactImportConfig, error) {
	// 🔹 Criamos um JSON de exemplo com os primeiros registros do CSV
	sampleData, _ := json.Marshal(sampleRecords[:3]) // Enviar 3 amostras para IA

	// 🔹 Criamos um exemplo de saída esperada
	expectedOutput := `
	{
		"about_data": {
			"source": "todos_os_campos",
			"rules": "Do que se trata a base de dados? Qual o contexto? Quais informações são relevantes?"
		},
		"name": {
			"source": "nome",
			"rules": "Utilizar o nome do contato"
		},
		"email": {
			"source": "email",
			"rules": "Utilizar o e-mail do contato"
		},
		"whatsapp": {
			"source": "fone_celular",
			"rules": "Se fone_celular não existir, verificar fone_residencial"
		},
		"gender": {
			"source": "nome",
			"rules": "Definir gênero pelo nome, se não for possível deixar vazio."
		},
		"birth_date": {
			"source": "data_nascimento",
			"rules": "Formatar para YYYY-MM-DD"
		},
		"bairro": {
			"source": "bairro",
			"rules": ""
		},
		"cidade": {
			"source": "cidade",
			"rules": ""
		},
		"estado": {
			"source": "uf",
			"rules": ""
		},
		"eventos": {
			"source": "cursos",
			"rules": "Associar cursos concluídos aos eventos relevantes"
		},
		"interesses": {
			"source": "cursos",
			"rules": "Relacionar cursos a interesses em áreas específicas. Possíveis categorias: marketing_vendas, tecnologia_da_informacao, design_multimidia, tecnicas_profissionais_manutencao, saude_bem_estar, idiomas, negocios_gestao_financas, desenvolvimento_pessoal_profissional, artesanato_moda, beleza_estetica, gastronomia_culinaria"
		},
		"perfil": {
			"source": "profissao,local_trabalho",
			"rules": "Categorizar perfil com base nas informações de profissão e local de trabalho. Possíveis categorias: industria, producao, construcao_civil, manutencao, logistica, comercial, tecnologia, saude_bem_estar, educacao, financas, gestao, marketing, seguranca, engenharia, juridico, agronegocio, meio_ambiente"
		},
		"history": {
			"source": "todos_os_campos",
			"rules": "Criar um breve histórico do aluno com base nas informações disponíveis. Máximo de 500 caracteres. Regras: 1. O nome do aluno deve estar capitalizado corretamente, no formato 'João da Silva'. 2. Datas devem ser formatadas como 'DD/MM/YYYY'. 3. O texto deve relatar apenas fatos de forma natural e profissional, evitando repetições desnecessárias."
		},
		"last_contact_at": {
			"source": "cursos",
			"rules": "Utilizar última data de conclusão de cursos. Formato: YYYY-MM-DD"
		}
	}
	`

	// 🔹 Criamos o prompt para IA
	prompt := fmt.Sprintf(`
		Estamos processando um CSV para importar contatos em um sistema CRM. Aqui estão os cabeçalhos do CSV:
		%s

		Abaixo estão algumas amostras dos dados reais do CSV:
		%s

		Queremos mapear esses dados para um sistema de CRM. 
		O resultado deve seguir o modelo JSON abaixo, onde "source" indica a coluna original e "rules" define regras adicionais:
		%s

		Por favor, retorne um JSON estruturado no mesmo formato.
	`, strings.Join(headers, ", "), sampleData, expectedOutput)

	// 🔹 Criamos um novo contexto com timeout de 60 segundos
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel() // Garante que o contexto será cancelado ao final da execução

	request := ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []ChatMessage{
			{Role: "system", Content: "Você é um assistente especializado em análise de dados para CRM. Você retorna exclusivamente JSON puro, sem marcações de código, sem comentários e sem texto adicional. Apenas retorne um objeto JSON válido."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.4,
	}

	aiResponse, err := s.openAIClient.CreateChatCompletion(ctxWithTimeout, request)
	if err != nil {
		return nil, fmt.Errorf("erro ao chamar OpenAI: %w", err)
	}

	if len(aiResponse.Choices) == 0 {
		return nil, fmt.Errorf("nenhuma resposta válida da OpenAI")
	}

	// Extrai o JSON da resposta
	cleanJSON := utils.SanitizeJSONResponse(aiResponse.Choices[0].Message.Content)

	var config models.ContactImportConfig
	if err := json.Unmarshal([]byte(cleanJSON), &config); err != nil {
		return nil, fmt.Errorf("erro ao converter JSON para DTO: %w", err)
	}

	s.log.Debug("Configuração de importação gerada pela OpenAI",
		slog.String("json", cleanJSON))

	return &config, nil
}
