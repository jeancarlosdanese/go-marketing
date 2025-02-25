// File: /internal/server/routes/campaign_settings_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterCampaignSettingsRoutes adiciona as rotas relacionadas às configurações das campanhas
func RegisterCampaignSettingsRoutes(
	mux *http.ServeMux,
	authMiddleware func(http.Handler) http.HandlerFunc,
	settingsRepo db.CampaignSettingsRepository,
) {
	handler := handlers.NewCampaignSettingsHandler(settingsRepo)

	// 📌 Criar configurações para uma campanha
	mux.Handle("POST /campaigns/{campaign_id}/settings", authMiddleware(handler.CreateSettingsHandler()))

	// 📌 Obter configurações de uma campanha
	mux.Handle("GET /campaigns/{campaign_id}/settings", authMiddleware(handler.GetSettingsHandler()))

	// 📌 Atualizar configurações de uma campanha
	mux.Handle("PUT /campaigns/{campaign_id}/settings", authMiddleware(handler.UpdateSettingsHandler()))

	// 📌 Excluir configurações de uma campanha
	mux.Handle("DELETE /campaigns/{campaign_id}/settings", authMiddleware(handler.DeleteSettingsHandler()))

	// 📌 Obter última configuração usada por uma conta (para reutilização)
	mux.Handle("POST /campaigns/{campaign_id}/clone-last-settings", authMiddleware(handler.GetLastSettingsHandler()))
}
