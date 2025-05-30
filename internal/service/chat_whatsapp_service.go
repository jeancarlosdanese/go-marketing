// internal/service/chat_whatsapp_service.go

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type ChatWhatsAppService interface {
	RegistrarChat(ctx context.Context, chat *models.Chat) (*models.Chat, error)
	ListarChatsPorConta(ctx context.Context, accountID uuid.UUID) ([]*models.Chat, error)
	BuscarChatPorID(ctx context.Context, accountID, chatID uuid.UUID) (*models.Chat, error)
	AtualizarChat(ctx context.Context, accountID, chatID uuid.UUID, data dto.ChatUpdateDTO) (*models.Chat, error)
	ListarContatosDoChat(ctx context.Context, accountID, chatID uuid.UUID) ([]dto.ChatContactFull, error)
	RegistrarMensagemManual(ctx context.Context, accountID, chatID, chatContactID uuid.UUID, chatMessage dto.ChatMessageCreateDTO) (*models.ChatMessage, error)
	ListarMensagens(ctx context.Context, accountID, chatID, chatContactID uuid.UUID) ([]models.ChatMessage, error)
	SugestaoRespostaAI(ctx context.Context, accountID, chatID, chatContactID uuid.UUID, message string) (string, error)
	ProcessarMensagemRecebida(ctx context.Context, webhookBaileysPayload *dto.WebhookBaileysPayload) error

	IniciarSessaoWhatsApp(ctx context.Context, accountID, chatID uuid.UUID) (*StartSessionResponse, error)
	ObterQRCodeSessao(ctx context.Context, accountID, chatID uuid.UUID) (*QRCodeResponse, error)
	VerificarSessionStatusViaAPI(ctx context.Context, accountID, chatID uuid.UUID) (*dto.SessionStatusDTO, error)
}

type chatWhatsAppService struct {
	log                 *slog.Logger
	chatRepo            db.ChatRepository
	contactRepo         db.ContactRepository
	whatsAppContactRepo db.WhatsappContactRepository
	chatContactRepo     db.ChatContactRepository
	chatMessageRepo     db.ChatMessageRepository
	openaiService       OpenAIService
	baileysService      WhatsAppBaileysService
	// evolutionService EvolutionService
}

func NewChatWhatsAppService(
	chatRepo db.ChatRepository,
	contactRepo db.ContactRepository,
	whatsAppContactRepo db.WhatsappContactRepository,
	chatContactRepo db.ChatContactRepository,
	chatMessageRepo db.ChatMessageRepository,
	openaiService OpenAIService,
	baileysService WhatsAppBaileysService,
	// evolution EvolutionService,
) ChatWhatsAppService {
	return &chatWhatsAppService{
		log:                 logger.GetLogger(),
		chatRepo:            chatRepo,
		contactRepo:         contactRepo,
		whatsAppContactRepo: whatsAppContactRepo,
		chatContactRepo:     chatContactRepo,
		chatMessageRepo:     chatMessageRepo,
		openaiService:       openaiService,
		baileysService:      baileysService,
		// evolutionService: evolution,
	}
}

func (s *chatWhatsAppService) RegistrarChat(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	result, err := s.chatRepo.Insert(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar chat: %w", err)
	}
	return result, nil
}

// SugestaoRespostaAI gera uma sugestão de resposta usando IA com base no histórico do chat
func (s *chatWhatsAppService) SugestaoRespostaAI(ctx context.Context, accountID, chatID, chatContactID uuid.UUID, message string) (string, error) {
	// 1. Buscar chat ativo do setor
	chat, err := s.chatRepo.GetActiveByID(ctx, accountID, chatID)
	if err != nil {
		return "", fmt.Errorf("chat não encontrado para o setor %s: %w", chatID, err)
	}

	// 2. Buscar relação chat_contact
	chatContact, err := s.chatContactRepo.FindByID(ctx, accountID, chatID, chatContactID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar ou criar chat_contact: %w", err)
	}

	// 3. Buscar dados do whatsapp contact
	whatsappContact, err := s.whatsAppContactRepo.FindByID(ctx, chatContact.WhatsappContactID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar contato do WhatsApp: %w", err)
	}

	// 4. Buscar contato no CRM
	contact, err := s.contactRepo.GetByID(ctx, whatsappContact.ContactID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar contato no CRM: %w", err)
	}

	// 5. Buscar mensagens anteriores do chat
	chatMessages, err := s.chatMessageRepo.ListByChatContact(ctx, chatContact.ID)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar mensagens do chat: %w", err)
	}

	// 4. Gerar sugestão via OpenAI
	prompt := buildPrompt(contact, chatMessages, message)
	resp, err := s.openaiService.CreateChatCompletion(ctx, ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: chat.Instructions},
			{Role: "user", Content: prompt},
		},
		Temperature: 0,
	})
	if err != nil {
		return "", fmt.Errorf("erro na IA: %w", err)
	}

	s.log.Debug("Prompt enviado para IA", slog.String("prompt", prompt))

	sugestao := resp.Choices[0].Message.Content

	return sugestao, nil
}

// ListarChatsPorConta retorna todos os chats de uma conta
func (s *chatWhatsAppService) ListarChatsPorConta(ctx context.Context, accountID uuid.UUID) ([]*models.Chat, error) {
	return s.chatRepo.ListByAccountID(ctx, accountID)
}

// BuscarChatPorID busca um chat por ID e conta
func (s *chatWhatsAppService) BuscarChatPorID(ctx context.Context, accountID, chatID uuid.UUID) (*models.Chat, error) {
	return s.chatRepo.GetByID(ctx, accountID, chatID)
}

func (s *chatWhatsAppService) AtualizarChat(ctx context.Context, accountID, chatID uuid.UUID, data dto.ChatUpdateDTO) (*models.Chat, error) {
	chat, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, err
	}

	// Atualiza os campos
	chat.Title = data.Title
	chat.Instructions = data.Instructions
	chat.PhoneNumber = data.PhoneNumber
	chat.InstanceName = data.InstanceName
	chat.WebhookURL = data.WebhookURL
	chat.UpdatedAt = time.Now()

	return s.chatRepo.Update(ctx, chat)
}

// buildPrompt constrói o prompt para a IA com base no contato, mensagens anteriores e nova mensagem
func buildPrompt(c *models.Contact, chatMessages []models.ChatMessage, message string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "📇 CONTATO:\n- Nome: %s\n", c.Name)
	if c.Cidade != nil && c.Estado != nil {
		fmt.Fprintf(&b, "- Localização: %s - %s\n", *c.Cidade, *c.Estado)
	}
	if c.History != nil {
		fmt.Fprintf(&b, "- Histórico: %s\n", *c.History)
	}
	if c.Tags != nil && len(c.Tags.Perfil) > 0 {
		perfis := make([]string, len(c.Tags.Perfil))
		for i, tag := range c.Tags.Perfil {
			perfis[i] = *tag
		}
		fmt.Fprintf(&b, "- Perfil: %s\n", strings.Join(perfis, ", "))
	}

	fmt.Fprintf(&b, "\n💬 CONVERSA ANTERIOR:\n")
	for _, msg := range chatMessages {
		autor := "Cliente"
		if msg.Actor == "atendente" {
			autor = "Atendente"
		}
		fmt.Fprintf(&b, "%s: %s\n", autor, msg.Content)
	}

	fmt.Fprintf(&b, "\n📥 MENSAGEM RECEBIDA:\nCliente: %s\n", message)

	fmt.Fprintf(&b, "\n📌 ORIENTAÇÃO:\nCom base nas informações acima, sugira uma próxima resposta adequada para o WhatsApp.\nA mensagem deve:\n")
	fmt.Fprintf(&b, "- Manter o tom alinhado com o atendimento comercial da Hyberica;\n")
	fmt.Fprintf(&b, "- Esclarecer dúvidas, apresentar soluções ou conduzir o diálogo conforme necessário;\n")
	fmt.Fprintf(&b, "- Ser objetiva, cordial e informativa.\n")

	return b.String()
}

// ListarContatosDoChat retorna todos os contatos de um chat com dados adicionais
func (s *chatWhatsAppService) ListarContatosDoChat(ctx context.Context, accountID, chatID uuid.UUID) ([]dto.ChatContactFull, error) {
	chatContacts, err := s.chatContactRepo.ListByChatID(ctx, accountID, chatID)
	if err != nil {
		return nil, fmt.Errorf("erro ao listar contatos do chat: %w", err)
	}

	return chatContacts, nil
}

// RegistrarMensagemManual registra uma mensagem manual no chat
func (s *chatWhatsAppService) RegistrarMensagemManual(ctx context.Context, accountID, chatID, chatContactID uuid.UUID, chatMessage dto.ChatMessageCreateDTO) (*models.ChatMessage, error) {
	// 🔹 Verifica se o chat é da conta
	chat, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, err
	}

	// 🔹 Obtém ou cria o relacionamento
	chatContact, err := s.chatContactRepo.FindByID(ctx, accountID, chatID, chatContactID)
	if err != nil {
		return nil, err
	}

	// 🔹 Cria nova mensagem
	msg := models.ChatMessage{
		ChatContactID:   chatContact.ID,
		Actor:           chatMessage.Actor,
		Type:            chatMessage.Type,
		Content:         chatMessage.Content,
		FileURL:         chatMessage.FileURL,
		SourceProcessed: false,
	}

	messageCreated, err := s.chatMessageRepo.Create(ctx, msg)
	if err != nil {
		return nil, fmt.Errorf("erro ao registrar mensagem: %w", err)
	}
	s.log.Debug("Mensagem registrada com sucesso", slog.Any("mensagem", messageCreated))

	// Após salvar a mensagem no banco
	if messageCreated.Actor == "atendente" || messageCreated.Actor == "ai" {
		// s.log.Debug("Mensagem registrada com sucesso", slog.String("mensagem", chatMessage.Content))
		// Buscar whatsapp contact pelo ID
		whatsappContact, err := s.whatsAppContactRepo.FindByID(ctx, chatContact.WhatsappContactID)
		if err == nil {
			// err = s.evolutionService.SendTextMessage(chat.InstanceName, *contato.WhatsApp, chatMessage.Content)
			_, err = s.baileysService.SendTextMessage(chat.InstanceName, whatsappContact.JID, messageCreated.Content)
			if err != nil {
				s.log.Error("Erro ao enviar mensagem para o WhatsApp", slog.String("numero", whatsappContact.Phone), slog.String("mensagem", messageCreated.Content), slog.Any("erro", err))
			} else {
				s.log.Debug("Mensagem enviada com sucesso para o WhatsApp", slog.String("numero", whatsappContact.Phone), slog.String("mensagem", messageCreated.Content))
			}
		}
	}

	return messageCreated, nil
}

// ListarMensagens retorna todas as mensagens de um chat
func (s *chatWhatsAppService) ListarMensagens(ctx context.Context, accountID, chatID, chatContactID uuid.UUID) ([]models.ChatMessage, error) {
	// 1. Retorna mensagens ordenadas
	return s.chatMessageRepo.ListByChatContact(ctx, chatContactID)
}

// ProcessarMensagemRecebida processa uma mensagem recebida do WhatsApp
func (s *chatWhatsAppService) ProcessarMensagemRecebida(ctx context.Context, webhookBaileysPayload *dto.WebhookBaileysPayload) error {
	// 🔹 1. Buscar o chat com base na instância da GetActiveByInstanceName
	chat, err := s.chatRepo.GetActiveByInstanceName(ctx, webhookBaileysPayload.SessionID)
	if err != nil {
		return fmt.Errorf("nenhum chat ativo com instance_name=%s: %w", webhookBaileysPayload.SessionID, err)
	}

	// 🔹 2. Extrair número do remoteJid (ex: 554999999999@...)
	normalizedNumber := utils.NormalizeWhatsAppNumber(webhookBaileysPayload.From)

	res, err := s.baileysService.ResolveNumber(webhookBaileysPayload.SessionID, normalizedNumber)
	if err != nil {
		s.log.Error("Erro ao resolver número via Baileys", slog.String("normalizedNumber", normalizedNumber), slog.Any("erro", err))
		return err
	}
	if res == nil {
		return fmt.Errorf("resposta nula da API para número %s", normalizedNumber)
	}
	if !res.Found {
		return fmt.Errorf("número %s não encontrado no WhatsApp", normalizedNumber)
	}

	// 🔹 3. Enriquecer dados com IA
	enrichedContact, err := s.EnriquecerContatoComIA(ctx, webhookBaileysPayload, res.BusinessProfile)
	if err != nil {
		s.log.Warn("IA falhou ao enriquecer contato, usando fallback", slog.Any("erro", err))
		// fallback mínimo
		enrichedContact = &models.Contact{
			Name:     webhookBaileysPayload.PushName,
			WhatsApp: &normalizedNumber, // Garante que o número normalizado seja usado
		}
	}

	// 🔹 4. Buscar ou criar o contato
	contact, err := s.contactRepo.FindOrCreateByWhatsApp(ctx, chat.AccountID, enrichedContact)
	if err != nil {
		return fmt.Errorf("erro ao buscar ou criar contato: %w", err)
	}

	whatsAppContact := &models.WhatsappContact{
		AccountID:       contact.AccountID,
		ContactID:       contact.ID,
		Name:            contact.Name,
		Phone:           normalizedNumber,
		JID:             webhookBaileysPayload.From,
		IsBusiness:      res.IsBusiness,
		BusinessProfile: res.BusinessProfile,
	}

	whatsAppContact, err = s.whatsAppContactRepo.FindOrCreate(ctx, whatsAppContact)
	if err != nil {
		return fmt.Errorf("erro ao buscar ou criar contato do WhatsApp: %w", err)
	}

	// 🔹 5. Obter ou criar o relacionamento chat_contact
	chatContact, err := s.chatContactRepo.FindOrCreate(ctx, chat.AccountID, chat.ID, whatsAppContact.ID)
	if err != nil {
		return fmt.Errorf("erro ao buscar ou criar chat_contact: %w", err)
	}

	// 🔹 6. Salvar a mensagem recebida
	if webhookBaileysPayload.Type == "text" {

		msg := models.ChatMessage{
			ChatContactID:   chatContact.ID,
			Actor:           "cliente",
			Type:            "texto",
			Content:         webhookBaileysPayload.Message,
			FileURL:         "",
			SourceProcessed: false,
		}

		messageCreated, err := s.chatMessageRepo.Create(ctx, msg)
		if err != nil {
			return fmt.Errorf("erro ao registrar mensagem recebida: %w", err)
		}
		s.log.Debug("Mensagem recebida registrada com sucesso", slog.Any("mensagem", messageCreated))
	} else if webhookBaileysPayload.Type == "image" || webhookBaileysPayload.Type == "video" {
		msg := models.ChatMessage{
			ChatContactID:   chatContact.ID,
			Actor:           "cliente",
			Type:            webhookBaileysPayload.Type,
			Content:         webhookBaileysPayload.Message,
			FileURL:         "", // TODO: URL do arquivo enviado
			SourceProcessed: false,
		}

		messageCreated, err := s.chatMessageRepo.Create(ctx, msg)
		if err != nil {
			return fmt.Errorf("erro ao registrar mensagem recebida: %w", err)
		}
		s.log.Debug("Mensagem recebida registrada com sucesso", slog.Any("mensagem", messageCreated))
	} else {
		s.log.Warn("Tipo de mensagem não suportado", slog.String("type", webhookBaileysPayload.Type))
		return fmt.Errorf("tipo de mensagem não suportado: %s", webhookBaileysPayload.Type)
	}

	return nil
}

// IniciarSessaoWhatsApp inicia uma sessão do WhatsApp usando a API Baileys
func (s *chatWhatsAppService) IniciarSessaoWhatsApp(ctx context.Context, accountID, chatID uuid.UUID) (*StartSessionResponse, error) {
	chat, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat não encontrado: %w", err)
	}

	s.log.Debug("Iniciando sessão do WhatsApp", slog.String("instance_name", chat.InstanceName), slog.String("webhook_url", chat.WebhookURL))
	// Verifica se a instância e o webhook estão configurados

	if chat.InstanceName == "" || chat.WebhookURL == "" {
		return nil, fmt.Errorf("chat está sem instance_name ou webhook_url configurado")
	}

	return s.baileysService.StartSession(chat.InstanceName, chat.WebhookURL)
}

// ObterQRCodeSessao obtém o QR Code para autenticação da sessão do WhatsApp
func (s *chatWhatsAppService) ObterQRCodeSessao(ctx context.Context, accountID, chatID uuid.UUID) (*QRCodeResponse, error) {
	chat, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat não encontrado: %w", err)
	}

	if chat.InstanceName == "" {
		return nil, fmt.Errorf("chat está sem instance_name configurado")
	}

	return s.baileysService.GetQRCode(chat.InstanceName)
}

// VerificarSessionStatusViaAPI consulta o status da sessão e atualiza no banco
func (s *chatWhatsAppService) VerificarSessionStatusViaAPI(ctx context.Context, accountID, chatID uuid.UUID) (*dto.SessionStatusDTO, error) {
	chat, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat não encontrado: %w", err)
	}

	if chat.InstanceName == "" {
		return nil, fmt.Errorf("chat está sem instance_name configurado")
	}

	sessionStatus, err := s.baileysService.GetSessionState(chat.InstanceName)
	if err != nil {
		return nil, fmt.Errorf("erro ao obter status da sessão: %w", err)
	}

	if chat.SessionStatus != sessionStatus.Status {
		err = s.chatRepo.UpdateSessionStatus(ctx, chat.ID, sessionStatus.Status)
		if err != nil {
			return nil, fmt.Errorf("erro ao atualizar status da sessão no banco: %w", err)
		}
	}

	return sessionStatus, nil
}

func (s *chatWhatsAppService) EnriquecerContatoComIA(
	ctx context.Context,
	payload *dto.WebhookBaileysPayload,
	businessProfile *models.BusinessProfile,
) (*models.Contact, error) {
	// Estrutura auxiliar para input do prompt
	input := map[string]interface{}{
		"phone":    payload.Phone,
		"pushName": payload.PushName,
	}

	if businessProfile != nil {
		input["business_profile"] = businessProfile
	}

	// Serializa os dados recebidos em JSON
	rawData, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar dados para IA: %w", err)
	}

	// Mensagens para a IA no formato Chat
	request := ChatCompletionRequest{
		Model: "gpt-4.1-nano",
		Messages: []ChatMessage{
			{
				Role: "system",
				Content: `
Você é um assistente que atua como um classificador e enriquecedor de dados de contatos para um sistema de CRM.

Sua função é analisar os dados recebidos de uma mensagem do WhatsApp (incluindo nome, número e opcionalmente perfil comercial) e retornar um objeto JSON que preencha corretamente os campos da struct abaixo:


// Contact representa um contato no sistema
type Contact struct {
	Name          string       'json:"name"'                    // Nome do contato (sem sufixos de marca ou empresa)
	Email         *string      'json:"email,omitempty"'         // Se informado no perfil comercial
	WhatsApp      *string      'json:"whatsapp,omitempty"'      // Apenas números com DDI (ex: 554991234567)
	Gender        *string      'json:"gender,omitempty"'        // "masculino", "feminino", "outro" — inferir apenas com alta confiança
	BirthDate     *time.Time   'json:"birth_date,omitempty"'    // Se conhecido, formato RFC3339
	Bairro        *string      'json:"bairro,omitempty"'        // Extrair do endereço (se aplicável)
	Cidade        *string      'json:"cidade,omitempty"'
	Estado        *string      'json:"estado,omitempty"'
	Tags          *ContactTags 'json:"tags,omitempty"'          // JSONB com campos interesses, perfil, eventos
	History       *string      'json:"history,omitempty"'       // Registrar primeira mensagem recebida com data
}

type ContactTags struct {
	Interesses []*string 'json:"interesses,omitempty"'
	Perfil     []*string 'json:"perfil,omitempty"'
	Eventos    []*string 'json:"eventos,omitempty"'
}

### INSTRUÇÕES:

1. Use os dados brutos fornecidos em JSON abaixo.
2. Quando 'business_profile' estiver presente, **use os campos 'category' e 'description'** para:
   - Preencher 'tags.perfil' ou 'tags.interesses' se aplicável.
   - Complementar o campo 'history', descrevendo o ramo da empresa.
3. Quando 'business_profile' **não estiver presente**, assuma que o remetente é uma pessoa física. Ignore qualquer dado do perfil comercial e baseie-se apenas nos campos 'pushName', 'phone'.
4. Campos opcionais ('email', 'bairro', 'tags', etc.) devem ser preenchidos **apenas se houver informação clara ou inferência altamente confiável**.
5. Use 'null' explicitamente nos campos omissos. Responda apenas com o JSON da struct 'Contact'.
`,
			},
			{
				Role:    "user",
				Content: string(rawData),
			},
		},
		Temperature: 0.2,
		ResponseFormat: &ResponseFormat{
			Type: "json_schema",
			JSONSchema: JSONSchemaSpec{
				Name: "EnrichedContact",
				Schema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name":     map[string]string{"type": "string"},
						"email":    map[string]string{"type": "string", "format": "email"},
						"whatsapp": map[string]string{"type": "string"},
						"gender":   map[string]string{"type": "string"},
						"birth_date": map[string]string{
							"type":   "string",
							"format": "date-time",
						},
						"bairro":  map[string]string{"type": "string"},
						"cidade":  map[string]string{"type": "string"},
						"estado":  map[string]string{"type": "string"},
						"history": map[string]string{"type": "string"},
						"tags": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"interesses": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"type": "string"},
								},
								"perfil": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"type": "string"},
								},
								"eventos": map[string]interface{}{
									"type":  "array",
									"items": map[string]string{"type": "string"},
								},
							},
						},
					},
					"required": []string{"name"},
				},
			},
		},
	}

	response, err := s.openaiService.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("erro ao gerar resposta com IA: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("resposta da IA vazia")
	}

	// Tenta deserializar a resposta da IA
	enriched := &models.Contact{}
	if err := json.Unmarshal([]byte(response.Choices[0].Message.Content), enriched); err != nil {
		return nil, fmt.Errorf("erro ao interpretar resposta da IA: %w", err)
	}

	return enriched, nil
}
