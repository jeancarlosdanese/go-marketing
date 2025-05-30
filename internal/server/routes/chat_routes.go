// internal/server/routes/chat_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// RegisterChatRoutes adiciona as rotas do módulo de atendimento inteligente (Copiloto)
func RegisterChatRoutes(
	mux *http.ServeMux,
	authMiddleware func(http.Handler) http.HandlerFunc,
	chatRepo db.ChatRepository,
	contactRepo db.ContactRepository,
	chatContactRepo db.ChatContactRepository,
	chatMessageRepo db.ChatMessageRepository,
	openAIService service.OpenAIService,
	chatService service.ChatWhatsAppService,
) {
	chatHandler := handlers.NewChatWhatsAppHandler(chatService)

	// 🔒 Protegido por autenticação
	mux.Handle("POST /chats", authMiddleware(chatHandler.CreateChat()))
	mux.Handle("GET /chats", authMiddleware(chatHandler.ListChats()))
	mux.Handle("GET /chats/{chat_id}", authMiddleware(chatHandler.GetChatByID()))
	mux.Handle("PUT /chats/{chat_id}", authMiddleware(chatHandler.UpdateChat()))

	mux.Handle("GET /chats/{chat_id}/status", authMiddleware(chatHandler.VerificarStatusSessao()))

	mux.Handle("GET /chats/{chat_id}/chat-contacts", authMiddleware(chatHandler.ListarContatosDoChat()))
	mux.Handle("POST /chats/{chat_id}/chat-contacts/{chat_contact_id}/messages", authMiddleware(chatHandler.RegistrarMensagem()))
	mux.Handle("GET /chats/{chat_id}/chat-contacts/{chat_contact_id}/messages", authMiddleware(chatHandler.ListarMensagens()))

	mux.Handle("POST /chats/{chat_id}/chat-contacts/{chat_contact_id}/suggestion-ai", authMiddleware(chatHandler.SugestaoRespostaAI()))

	mux.Handle("POST /chats/{chat_id}/session-start", authMiddleware(chatHandler.IniciarSessaoWhatsApp()))
	mux.Handle("GET /chats/{chat_id}/qrcode", authMiddleware(chatHandler.ObterQrCodeHandler()))
}
