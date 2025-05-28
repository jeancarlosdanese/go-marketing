// File: /internal/server/routes/routes.go

package routes

import (
	"net/http"
	"os"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// NewRouter cria e retorna um roteador HTTP configurado.
func NewRouter(
	otpRepo db.AccountOTPRepository,
	accountRepo db.AccountRepository,
	accountSettingsRepo db.AccountSettingsRepository,
	contactRepo db.ContactRepository,
	templateRepo db.TemplateRepository,
	campaignRepo db.CampaignRepository,
	audienceRepo db.CampaignAudienceRepository,
	campaignSettingsRepo db.CampaignSettingsRepository,
	openAIService service.OpenAIService,
	campaignProcessor service.CampaignProcessorService,
	contactImportRepo db.ContactImportRepository,
	campaignMessageRepo db.CampaignMessageRepository,
	chatRepo db.ChatRepository,
	chatContactRepo db.ChatContactRepository,
	chatMessageRepo db.ChatMessageRepository,
) *http.ServeMux {
	mux := http.NewServeMux()

	// 🔥 Criar middlewares
	authMiddleware := middleware.AuthMiddleware(accountRepo)

	// 🔥 Registrar rotas principais
	RegisterAuthRoutes(mux, authMiddleware, otpRepo)
	RegisterAccountRoutes(mux, authMiddleware, accountRepo)
	RegisterAccountSettingsRoutes(mux, authMiddleware, accountSettingsRepo)
	RegisterContactRoutes(mux, authMiddleware, contactRepo, contactImportRepo, openAIService)
	RegisterTemplateRoutes(mux, authMiddleware, templateRepo)
	RegisterCampaignRoutes(mux, authMiddleware, campaignRepo, audienceRepo, campaignProcessor)
	RegisterCampaignAudienceRoutes(mux, authMiddleware, campaignRepo, contactRepo, audienceRepo)
	RegisterSESFeedBackRoutes(mux, audienceRepo, contactRepo)
	RegisterCampaignSettingsRoutes(mux, authMiddleware, campaignRepo, campaignSettingsRepo)
	RegisterCampaignMessageRoutes(mux, authMiddleware, campaignRepo, campaignSettingsRepo, contactRepo, audienceRepo, campaignMessageRepo, campaignProcessor)

	// 🔥 Registrar rotas do WhatsApp
	// evolutionService := service.NewEvolutionService()
	baileysService := service.NewWhatsAppBaileysService(os.Getenv("WHATSAPP_API_URL"), os.Getenv("WHATSAPP_API_KEY"))
	chatService := service.NewChatWhatsAppService(chatRepo, contactRepo, chatContactRepo, chatMessageRepo, openAIService, baileysService)
	RegisterChatRoutes(mux, authMiddleware, chatRepo, contactRepo, chatContactRepo, chatMessageRepo, openAIService, chatService)
	RegisterWebhookRoutes(mux, chatService)

	// 🔥 Rota de Health Check
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))

	return mux
}
