// File: /internal/server/routes/campaign_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// RegisterCampaignRoutes adiciona as rotas relacionadas às campanhas
func RegisterCampaignRoutes(
	mux *http.ServeMux,
	authMiddleware func(http.Handler) http.HandlerFunc,
	campaignRepo db.CampaignRepository,
	audienceRepo db.CampaignAudienceRepository,
	campaignProcessor service.CampaignProcessorService,
) {

	handler := handlers.NewCampaignHandle(campaignRepo, audienceRepo, campaignProcessor)

	// Criar campanha
	mux.Handle("POST /campaigns", authMiddleware(handler.CreateCampaignHandler()))

	// Listar todas as campanhas da conta autenticada
	mux.Handle("GET /campaigns", authMiddleware(handler.GetAllCampaignsHandler()))

	// Buscar uma campanha específica
	mux.Handle("GET /campaigns/{campaign_id}", authMiddleware(handler.GetCampaignHandler()))

	// Atualizar uma campanha
	mux.Handle("PUT /campaigns/{campaign_id}", authMiddleware(handler.UpdateCampaignHandler()))

	// Consultar status de uma campanha
	mux.Handle("GET /campaigns/{campaign_id}/status", authMiddleware(handler.GetCampaignStatusHandler()))

	// Atualizar status da campanha
	mux.Handle("PATCH /campaigns/{campaign_id}/status", authMiddleware(handler.UpdateCampaignStatusHandler()))

	// Deletar uma campanha
	mux.Handle("DELETE /campaigns/{campaign_id}", authMiddleware(handler.DeleteCampaignHandler()))
}
