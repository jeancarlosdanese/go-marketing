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
	SugerirResposta() http.HandlerFunc
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

type SugerirRespostaRequest struct {
	ChatID    string `json:"chat_id"`
	ContactID string `json:"contact_id"`
	Mensagem  string `json:"mensagem"`
}

type SugerirRespostaResponse struct {
	RespostaSugerida string `json:"resposta_sugerida"`
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

		lista, err := h.chatWhatsAppService.ListarContatosComDados(ctx, authAccount.ID, chatID)
		if err != nil {
			utils.SendError(w, 500, "Erro ao listar contatos do chat")
			h.log.Error("Erro ao listar contatos do chat", slog.Any("err", err))
			return
		}

		utils.SendSuccess(w, 200, lista)
	}
}

// RegistrarMensagem registra uma mensagem manualmente
func (h *chatWhatsAppHandler) RegistrarMensagem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		auth := middleware.GetAuthAccountOrFail(ctx, w, h.log)

		chatID := utils.GetUUIDFromRequestPath(r, w, "chat_id")
		contactID := utils.GetUUIDFromRequestPath(r, w, "contact_id")

		var req dto.ChatMessageCreateDTO
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, 400, "JSON inválido")
			return
		}

		msg, err := h.chatWhatsAppService.RegistrarMensagemManual(ctx, auth.ID, chatID, contactID, req)
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
		contactID := utils.GetUUIDFromRequestPath(r, w, "contact_id")

		mensagens, err := h.chatWhatsAppService.ListarMensagens(ctx, auth.ID, chatID, contactID)
		if err != nil {
			utils.SendError(w, 500, "Erro ao listar mensagens")
			return
		}

		utils.SendSuccess(w, 200, mensagens)
	}
}

// SugerirResposta gera uma resposta sugerida para a mensagem recebida
func (h *chatWhatsAppHandler) SugerirResposta() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		var req SugerirRespostaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.SendError(w, http.StatusBadRequest, "JSON inválido")
			return
		}

		h.log.Debug("request sugerir resposta", slog.String("chat_id", req.ChatID), slog.String("contact_id", req.ContactID), slog.String("mensagem", req.Mensagem))

		chatID, err := uuid.Parse(req.ChatID)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "chat_id inválido")
			return
		}
		contactID, err := uuid.Parse(req.ContactID)
		if err != nil {
			utils.SendError(w, http.StatusBadRequest, "contact_id inválido")
			return
		}

		resposta, err := h.chatWhatsAppService.SugerirRespostaIA(r.Context(), authAccount.ID, chatID, contactID, req.Mensagem)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao gerar resposta com IA")
			h.log.Error("erro ao gerar resposta com IA",
				slog.Any("err", err),
				slog.String("chat_id", chatID.String()),
				slog.String("contact_id", contactID.String()),
			)
			return
		}

		json.NewEncoder(w).Encode(SugerirRespostaResponse{
			RespostaSugerida: resposta,
		})
	}
}
