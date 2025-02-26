// File: /internal/server/routes/routes.go

package routes

import (
	"net/http"

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
) *http.ServeMux {
	mux := http.NewServeMux()

	// 🔥 Criar middlewares
	authMiddleware := middleware.AuthMiddleware(accountRepo)

	// 🔥 Registrar rotas principais
	RegisterAuthRoutes(mux, otpRepo)
	RegisterAccountRoutes(mux, authMiddleware, accountRepo)
	RegisterAccountSettingsRoutes(mux, authMiddleware, accountSettingsRepo)
	RegisterContactRoutes(mux, authMiddleware, contactRepo, openAIService)
	RegisterTemplateRoutes(mux, authMiddleware, templateRepo)
	RegisterCampaignRoutes(mux, authMiddleware, campaignRepo, audienceRepo, campaignProcessor)
	RegisterCampaignAudienceRoutes(mux, authMiddleware, campaignRepo, contactRepo, audienceRepo)
	RegisterSESFeedBackRoutes(mux, audienceRepo)
	RegisterCampaignSettingsRoutes(mux, authMiddleware, campaignSettingsRepo)

	// 🔥 Rota de Health Check
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))

	return mux
}
