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

	mux.Handle("GET /chats/{chat_id}/contatos", authMiddleware(chatHandler.ListarContatosDoChat()))
	mux.Handle("POST /chats/{chat_id}/contatos/{contact_id}/mensagens", authMiddleware(chatHandler.RegistrarMensagem()))
	mux.Handle("GET /chats/{chat_id}/contatos/{contact_id}/mensagens", authMiddleware(chatHandler.ListarMensagens()))

	mux.Handle("POST /chat/sugerir-resposta", authMiddleware(chatHandler.SugerirResposta()))

	mux.Handle("POST /chats/{chat_id}/iniciar-sessao", authMiddleware(chatHandler.IniciarSessaoWhatsApp()))
	mux.Handle("GET /chats/{chat_id}/qrcode", authMiddleware(chatHandler.ObterQrCodeHandler()))
}
