// internal/server/handlers/chat_whatsapp_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type ChatWhatsAppHandler interface {
	CreateChat() http.HandlerFunc
	ListChats() http.HandlerFunc
	GetChatByID() http.HandlerFunc
	UpdateChat() http.HandlerFunc
	ListarContatosDoChat() http.HandlerFunc
	RegistrarMensagem() http.HandlerFunc
	ListarMensagens() http.HandlerFunc
	SugestaoRespostaAI() http.HandlerFunc
	IniciarSessaoWhatsApp() http.HandlerFunc
	ObterQrCodeHandler() http.HandlerFunc
	VerificarStatusSessao() http.HandlerFunc
}

type chatWhatsAppHandler struct {
	log                 *slog.Logger
	chatWhatsAppService service.ChatWhatsAppService
}

func NewChatWhatsAppHandler(chatWhatsAppService service.ChatWhatsAppService) ChatWhatsAppHandler {
	return &chatWhatsAppHandler{
		log:                 logger.GetLogger(),
		chatWhatsAppService: chatWhatsAppService,
	}
}

type SuggestionRequest struct {
	Message string `json:"message"`
}

type SuggestionResponse struct {
	SuggestionAI string `json:"suggestion_ai"`
}

// CreateChat cria um novo chat
func (h *chatWhatsAppHandler) CreateChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authAccount := middleware.GetAuthAccountOrFail(ctx, w, h.log)

		var req dto.ChatCreateDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, 400, "JSON inválido")
			return
		}

		if err := req.Validate(); err != nil {
			utils.SendError(w, 422, err.Error())
			return
		}

		chat := req.ToModel()
		chat.AccountID = authAccount.ID

		created, err := h.chatWhatsAppService.RegistrarChat(ctx, chat)
		if err != nil {
			utils.SendError(w, 500, "Erro ao criar chat")
			h.log.Error("Erro ao registrar chat", slog.Any("err", err))
			return
		}

		utils.SendSuccess(w, 201, created)
	}
}

// ListChats retorna a lista de chats
func (h *chatWhatsAppHandler) ListChats() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authAccount := middleware.GetAuthAccountOrFail(ctx, w, h.log)

		chats, err := h.chatWhatsAppService.ListarChatsPorConta(ctx, authAccount.ID)
		if err != nil {
			utils.SendError(w, 500, "Erro ao listar chats")
			h.log.Error("Erro ao listar chats", slog.Any("err", err))
			return
		}

		utils.SendSuccess(w, 200, chats)
	}
}

// GetChatByID retorna o chat com o ID especificado
func (h *chatWhatsAppHandler) GetChatByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authAccount := middleware.GetAuthAccountOrFail(ctx, w, h.log)

		// 🔍 Capturar `campaign_id` da URL
		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")

		chat, err := h.chatWhatsAppService.BuscarChatPorID(ctx, authAccount.ID, chatID)
		if err != nil {
			utils.SendError(w, 404, "Chat não encontrado")
			h.log.Warn("Chat não encontrado", slog.String("chat_id", chatID.String()))
			return
		}

		utils.SendSuccess(w, 200, chat)
	}
}

// UpdateChat atualiza o chat com o ID especificado
func (h *chatWhatsAppHandler) UpdateChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authAccount := middleware.GetAuthAccountOrFail(ctx, w, h.log)
		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")

		var req dto.ChatUpdateDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, 400, "JSON inválido")
			return
		}

		if err := req.Validate(); err != nil {
			utils.SendError(w, 422, err.Error())
			return
		}

		updated, err := h.chatWhatsAppService.AtualizarChat(ctx, authAccount.ID, chatID, req)
		if err != nil {
			utils.SendError(w, 500, "Erro ao atualizar chat")
			h.log.Error("Erro ao atualizar chat", slog.Any("err", err))
			return
		}

		utils.SendSuccess(w, 200, updated)
	}
}

// ListarContatosDoChat retorna todos os contatos de um chat
func (h *chatWhatsAppHandler) ListarContatosDoChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authAccount := middleware.GetAuthAccountOrFail(ctx, w, h.log)
		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")

		chatContacts, err := h.chatWhatsAppService.ListarContatosDoChat(ctx, authAccount.ID, chatID)
		if err != nil {
			utils.SendError(w, 500, "Erro ao listar contatos do chat")
			h.log.Error("Erro ao listar contatos do chat", slog.Any("err", err), slog.String("chat_id", chatID.String()))
			return
		}
		if len(chatContacts) == 0 {
			utils.SendSuccess(w, 200, []dto.ChatContactFull{})
			return
		}

		// 🔹 Retorna contatos do chat
		utils.SendSuccess(w, 200, chatContacts)
	}
}

// RegistrarMensagem registra uma mensagem manualmente
func (h *chatWhatsAppHandler) RegistrarMensagem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := middleware.GetAuthAccountOrFail(ctx, w, h.log)

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		chatContactID := utils.GetUUIDFromRequestPath(r, w, "chat_contact_id")

		var req dto.ChatMessageCreateDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, 400, "JSON inválido")
			return
		}

		msg, err := h.chatWhatsAppService.RegistrarMensagemManual(ctx, auth.ID, chatID, chatContactID, req)
		if err != nil {
			utils.SendError(w, 500, "Erro ao registrar mensagem")
			return
		}

		utils.SendSuccess(w, 201, msg)
	}
}

// ListarMensagens retorna todas as mensagens de um contato específico
func (h *chatWhatsAppHandler) ListarMensagens() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := middleware.GetAuthAccountOrFail(ctx, w, h.log)

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		chatContactID := utils.GetUUIDFromRequestPath(r, w, "chat_contact_id")

		mensagens, err := h.chatWhatsAppService.ListarMensagens(ctx, auth.ID, chatID, chatContactID)
		if err != nil {
			utils.SendError(w, 500, "Erro ao listar mensagens")
			return
		}

		utils.SendSuccess(w, 200, mensagens)
	}
}

// SugestaoRespostaAI gera uma resposta sugerida para a mensagem recebida
func (h *chatWhatsAppHandler) SugestaoRespostaAI() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		chatContactID := utils.GetUUIDFromRequestPath(r, w, "chat_contact_id")

		var req SuggestionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, http.StatusBadRequest, "JSON inválido")
			return
		}

		h.log.Debug("request sugerir resposta", slog.String("chat_id", chatID.String()), slog.String("chat_contact_id", chatContactID.String()), slog.String("mensagem", req.Message))

		resposta, err := h.chatWhatsAppService.SugestaoRespostaAI(r.Context(), authAccount.ID, chatID, chatContactID, req.Message)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao gerar resposta com IA")
			h.log.Error("erro ao gerar resposta com IA",
				slog.Any("err", err),
				slog.String("chat_id", chatID.String()),
				slog.String("contact_id", chatContactID.String()),
			)
			return
		}

		json.NewEncoder(w).Encode(SuggestionResponse{
			SuggestionAI: resposta,
		})
	}
}

// IniciarSessaoWhatsApp inicia uma sessão do WhatsApp via ChatWhatsAppService
func (h *chatWhatsAppHandler) IniciarSessaoWhatsApp() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authAccount := middleware.GetAuthAccountOrFail(ctx, w, h.log)
		if authAccount == nil {
			return
		}

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		if chatID == uuid.Nil {
			return
		}

		resp, err := h.chatWhatsAppService.IniciarSessaoWhatsApp(ctx, authAccount.ID, chatID)
		if err != nil {
			h.log.Error("Erro ao iniciar sessão WhatsApp", slog.Any("erro", err))
			utils.SendError(w, 500, err.Error())
			return
		}

		json.NewEncoder(w).Encode(resp)
	}
}

// ObterQrCodeHandler obtém o QR Code para autenticação do WhatsApp
func (h *chatWhatsAppHandler) ObterQrCodeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		account := middleware.GetAuthAccountOrFail(r.Context(), w, logger.GetLogger())
		if account == nil {
			return
		}

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		if chatID == uuid.Nil {
			return
		}

		result, err := h.chatWhatsAppService.ObterQRCodeSessao(r.Context(), account.ID, chatID)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		json.NewEncoder(w).Encode(result)
	}
}

// VerificarStatusSessao verifica o status da sessão do WhatsApp
func (h *chatWhatsAppHandler) VerificarStatusSessao() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		account := middleware.GetAuthAccountOrFail(ctx, w, h.log)
		if account == nil {
			return
		}

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		if chatID == uuid.Nil {
			return
		}

		// 🔹 Consulta status no whatsapp-api
		sessionStatus, err := h.chatWhatsAppService.VerificarSessionStatusViaAPI(ctx, account.ID, chatID)
		if err != nil {
			h.log.Error("Erro ao verificar status no whatsapp-api", slog.String("erro", err.Error()))
			utils.SendError(w, 500, "Erro ao verificar status")
			return
		}

		// 🔹 Retorna status atualizado
		utils.SendSuccess(w, 200, sessionStatus)
	}
}
