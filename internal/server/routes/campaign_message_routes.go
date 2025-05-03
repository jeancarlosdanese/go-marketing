// internal/server/routes/campaign_message_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// RegisterCampaignMessageRoutes adiciona rotas relacionadas à geração de mensagens por IA
func RegisterCampaignMessageRoutes(
	mux *http.ServeMux,
	authMiddleware func(http.Handler) http.HandlerFunc,
	campaignRepo db.CampaignRepository,
	settingsRepo db.CampaignSettingsRepository,
	contactRepo db.ContactRepository,
	audienceRepo db.CampaignAudienceRepository,
	messageRepo db.CampaignMessageRepository,
	processor service.CampaignProcessorService,
) {
	handler := handlers.NewCampaignMessageHandler(campaignRepo, settingsRepo, contactRepo, audienceRepo, messageRepo, processor)

	mux.Handle("POST /campaigns/{campaign_id}/generate-message", authMiddleware(handler.GenerateCampaignMessageHandler()))
}
