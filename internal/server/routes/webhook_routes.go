// internal/server/routes/webhook_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// RegisterWebhookRoutes registra o endpoint público do webhook
func RegisterWebhookRoutes(mux *http.ServeMux, chatService service.ChatWhatsAppService) {
	webhookHandler := handlers.NewWebhookHandler(chatService)

	// 🔓 Webhook é público — não passa por authMiddleware
	mux.Handle("POST /webhook", webhookHandler.Handle())
}
