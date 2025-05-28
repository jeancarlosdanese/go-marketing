// internal/service/chat_whatsapp_service.go

package service

import (
	"context"
	"fmt"
	"log/slog"
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
	ListarContatosDoChat(ctx context.Context, accountID, chatID uuid.UUID) ([]*models.ChatContact, error)
	ListarContatosComDados(ctx context.Context, accountID, chatID uuid.UUID) ([]dto.ChatContactResumoDTO, error)
	RegistrarMensagemManual(ctx context.Context, accountID, chatID, contactID uuid.UUID, chatMessage dto.ChatMessageCreateDTO) (*models.ChatMessage, error)
	ListarMensagens(ctx context.Context, accountID, chatID, contactID uuid.UUID) ([]*models.ChatMessage, error)
	SugerirRespostaIA(ctx context.Context, accountID, chatID, contactID uuid.UUID, mensagemRecebida string) (string, error)
	ProcessarMensagemRecebida(ctx context.Context, instanceName, remoteJid, texto string) error

	IniciarSessaoWhatsApp(ctx context.Context, accountID, chatID uuid.UUID) (*StartSessionResponse, error)
	ObterQRCodeSessao(ctx context.Context, accountID, chatID uuid.UUID) (*QRCodeResponse, error)
	VerificarSessionStatusViaAPI(ctx context.Context, accountID, chatID uuid.UUID) (*dto.SessionStatusDTO, error)
}

type chatWhatsAppService struct {
	log             *slog.Logger
	chatRepo        db.ChatRepository
	contactRepo     db.ContactRepository
	chatContactRepo db.ChatContactRepository
	chatMessageRepo db.ChatMessageRepository
	openaiService   OpenAIService
	baileysService  WhatsAppBaileysService
	// evolutionService EvolutionService
}

func NewChatWhatsAppService(
	chatRepo db.ChatRepository,
	contactRepo db.ContactRepository,
	chatContactRepo db.ChatContactRepository,
	chatMessageRepo db.ChatMessageRepository,
	openaiService OpenAIService,
	baileysService WhatsAppBaileysService,
	// evolution EvolutionService,
) ChatWhatsAppService {
	return &chatWhatsAppService{
		log:             logger.GetLogger(),
		chatRepo:        chatRepo,
		contactRepo:     contactRepo,
		chatContactRepo: chatContactRepo,
		chatMessageRepo: chatMessageRepo,
		openaiService:   openaiService,
		baileysService:  baileysService,
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

func (s *chatWhatsAppService) SugerirRespostaIA(ctx context.Context, accountID, chatID, contactID uuid.UUID, mensagemRecebida string) (string, error) {
	// 1. Buscar chat ativo do setor
	chat, err := s.chatRepo.GetActiveByID(ctx, accountID, chatID)
	if err != nil {
		return "", fmt.Errorf("chat não encontrado para o setor %s: %w", chatID, err)
	}

	// 2. Buscar ou criar o relacionamento com o contato
	_, err = s.chatContactRepo.FindOrCreate(ctx, accountID, chat.ID, contactID)
	if err != nil {
		return "", fmt.Errorf("erro ao obter chat_contact: %w", err)
	}

	// 3. Buscar dados do contato
	contact, err := s.contactRepo.GetByID(ctx, contactID)
	if err != nil {
		return "", fmt.Errorf("contato não encontrado: %w", err)
	}

	// 4. Gerar sugestão via OpenAI
	prompt := buildPrompt(chat.Instructions, contact, mensagemRecebida)
	resp, err := s.openaiService.CreateChatCompletion(ctx, ChatCompletionRequest{
		Model: "gpt-4o-mini",
		Messages: []ChatMessage{
			{Role: "system", Content: "Você é um assistente de atendimento do setor " + chat.Department},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.4,
	})
	if err != nil {
		return "", fmt.Errorf("erro na IA: %w", err)
	}

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

// ListarContatosDoChat retorna todos os contatos de um chat
func (s *chatWhatsAppService) ListarContatosDoChat(ctx context.Context, accountID, chatID uuid.UUID) ([]*models.ChatContact, error) {
	// Verifica se o chat pertence à conta
	_, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, fmt.Errorf("chat não encontrado ou não pertence à conta: %w", err)
	}

	return s.chatContactRepo.ListByChatID(ctx, accountID, chatID)
}

func buildPrompt(instructions string, c *models.Contact, mensagem string) string {
	cidade := ""
	if c.Cidade != nil {
		cidade = *c.Cidade
	}
	estado := ""
	if c.Estado != nil {
		estado = *c.Estado
	}
	history := ""
	if c.History != nil {
		history = *c.History
	}
	return fmt.Sprintf(`
Mensagem recebida do cliente: "%s"

Informações do contato:
- Nome: %s
- Cidade: %s - %s
- Histórico: %s

Instruções do setor:
%s

Gere uma resposta adequada em tom cordial e objetivo.
Nunca envie chave PIX real. Use o marcador [INSERIR CHAVE PIX AQUI] se necessário.
`, mensagem, c.Name, cidade, estado, history, instructions)
}

// ListarContatosComDados retorna todos os contatos de um chat com dados adicionais
func (s *chatWhatsAppService) ListarContatosComDados(ctx context.Context, accountID, chatID uuid.UUID) ([]dto.ChatContactResumoDTO, error) {
	contatos, err := s.chatContactRepo.ListByChatID(ctx, accountID, chatID)
	if err != nil {
		return nil, err
	}

	var dtos []dto.ChatContactResumoDTO

	for _, c := range contatos {
		contact, err := s.contactRepo.GetByID(ctx, uuid.MustParse(c.ContactID))
		if err != nil {
			// Ignora contatos órfãos ou removidos
			continue
		}

		dtos = append(dtos, dto.ChatContactResumoDTO{
			ID:         c.ID,
			ContactID:  c.ContactID,
			Nome:       contact.Name,
			WhatsApp:   *contact.WhatsApp,
			Status:     c.Status,
			Atualizado: c.UpdatedAt,
		})
	}

	return dtos, nil
}

// RegistrarMensagemManual registra uma mensagem manual no chat
func (s *chatWhatsAppService) RegistrarMensagemManual(ctx context.Context, accountID, chatID, contactID uuid.UUID, chatMessage dto.ChatMessageCreateDTO) (*models.ChatMessage, error) {
	// 🔹 Verifica se o chat é da conta
	chat, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, err
	}

	// 🔹 Obtém ou cria o relacionamento
	chatContact, err := s.chatContactRepo.FindOrCreate(ctx, accountID, chatID, contactID)
	if err != nil {
		return nil, err
	}

	// 🔹 Cria nova mensagem
	msg := models.ChatMessage{
		ID:              uuid.NewString(),
		ChatContactID:   chatContact.ID,
		Actor:           chatMessage.Actor,
		Type:            chatMessage.Type,
		Content:         chatMessage.Content,
		FileURL:         chatMessage.FileURL,
		SourceProcessed: false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err = s.chatMessageRepo.Insert(ctx, msg)
	if err != nil {
		return nil, err
	}

	// Após salvar a mensagem no banco
	if chatMessage.Actor == "atendente" || chatMessage.Actor == "ai" {
		// s.log.Debug("Mensagem registrada com sucesso", slog.String("mensagem", chatMessage.Content))
		// Buscar contato
		contato, err := s.contactRepo.GetByID(ctx, contactID)
		if err == nil && contato.WhatsApp != nil {
			// err = s.evolutionService.SendTextMessage(chat.InstanceName, *contato.WhatsApp, chatMessage.Content)
			_, err = s.baileysService.SendTextMessage(chat.InstanceName, *contato.WhatsApp, chatMessage.Content)
			if err != nil {
				s.log.Error("Erro ao enviar mensagem para o WhatsApp", slog.String("numero", *contato.WhatsApp), slog.String("mensagem", chatMessage.Content), slog.Any("erro", err))
			} else {
				s.log.Debug("Mensagem enviada com sucesso para o WhatsApp", slog.String("numero", *contato.WhatsApp), slog.String("mensagem", chatMessage.Content))
			}
		}
	}

	return &msg, nil
}

// ListarMensagens retorna todas as mensagens de um chat
func (s *chatWhatsAppService) ListarMensagens(ctx context.Context, accountID, chatID, contactID uuid.UUID) ([]*models.ChatMessage, error) {
	// 1. Busca ou valida o chat
	_, err := s.chatRepo.GetByID(ctx, accountID, chatID)
	if err != nil {
		return nil, err
	}

	// 2. Busca relação chat_contact
	chatContact, err := s.chatContactRepo.FindOrCreate(ctx, accountID, chatID, contactID)
	if err != nil {
		return nil, err
	}

	// 3. Retorna mensagens ordenadas
	return s.chatMessageRepo.ListByChatContact(ctx, uuid.MustParse(chatContact.ID))
}

// ProcessarMensagemRecebida processa uma mensagem recebida do WhatsApp
func (s *chatWhatsAppService) ProcessarMensagemRecebida(ctx context.Context, instanceName, remoteJid, texto string) error {
	// 🔹 1. Buscar o chat com base na instância da GetActiveByInstanceName
	chat, err := s.chatRepo.GetActiveByInstanceName(ctx, instanceName)
	if err != nil {
		return fmt.Errorf("nenhum chat ativo com instance_name=%s: %w", instanceName, err)
	}

	// 🔹 2. Extrair número do remoteJid (ex: 554999999999@...)
	numero := utils.ExtractWhatsAppNumber(remoteJid)
	if numero == "" {
		return fmt.Errorf("número inválido extraído de: %s", remoteJid)
	}

	// 🔹 3. Buscar ou criar o contato
	contact, err := s.contactRepo.FindOrCreateByWhatsApp(ctx, chat.AccountID, numero)
	if err != nil {
		return fmt.Errorf("erro ao buscar ou criar contato: %w", err)
	}

	// 🔹 4. Obter ou criar o relacionamento chat_contact
	chatContact, err := s.chatContactRepo.FindOrCreate(ctx, chat.AccountID, chat.ID, contact.ID)
	if err != nil {
		return fmt.Errorf("erro ao buscar ou criar chat_contact: %w", err)
	}

	// 🔹 5. Salvar a mensagem recebida
	now := time.Now()
	msg := models.ChatMessage{
		ID:              uuid.NewString(),
		ChatContactID:   chatContact.ID,
		Actor:           "cliente",
		Type:            "texto",
		Content:         texto,
		FileURL:         "",
		SourceProcessed: false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.chatMessageRepo.Insert(ctx, msg); err != nil {
		return fmt.Errorf("erro ao salvar mensagem: %w", err)
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
