// File: /internal/server/routes/ses_feedback_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterSESFeedBackRoutes adiciona as rotas relacionadas à audiência de campanhas
func RegisterSESFeedBackRoutes(mux *http.ServeMux, audienceRepo db.CampaignAudienceRepository) {

	handler := handlers.NewSESFeedbackHandler(audienceRepo)

	// 📌 Atualiza contactAudience
	mux.Handle("POST /ses-feedback", handler.HandleSESFeedback())

	// SNS confirmend subscription
	// mux.Handle("POST /ses-feedback", handler.HandleSNSEvent())
}
