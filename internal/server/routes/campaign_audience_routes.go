// File: /internal/server/routes/campaign_audience_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterCampaignAudienceRoutes adiciona as rotas relacionadas à audiência de campanhas
func RegisterCampaignAudienceRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, campaignRepo db.CampaignRepository, contactRepo db.ContactRepository, audienceRepo db.CampaignAudienceRepository) {

	handler := handlers.NewCampaignAudienceHandle(campaignRepo, contactRepo, audienceRepo)

	// 📌 Adicionar contatos a uma campanha
	mux.Handle("POST /campaigns/{campaign_id}/audience", authMiddleware(handler.AddContactsToCampaignHandler()))

	// 📌 Obter audiência de uma campanha
	mux.Handle("GET /campaigns/{campaign_id}/audience", authMiddleware(handler.GetCampaignAudienceHandler()))
}
