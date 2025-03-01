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

	// 📌 Obter contatos disponíveis para uma campanha
	mux.Handle("GET /campaigns/{campaign_id}/available-contacts", authMiddleware(handler.GetAvailableContactsHandler()))

	// 📌 Obter audiência de uma campanha
	mux.Handle("GET /campaigns/{campaign_id}/audience", authMiddleware(handler.GetPaginatedCampaignAudienceHandler()))

	// 📌 Delete audiência de uma campanha
	mux.Handle("DELETE /campaigns/{campaign_id}/audience/{audience_id}", authMiddleware(handler.RemoveContactFromCampaignHandler()))

}
