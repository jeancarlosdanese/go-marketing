// File: /internal/server/routes/template_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterTemplateRoutes adiciona as rotas relacionadas a templates
func RegisterTemplateRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, templateRepo db.TemplateRepository) {
	handler := handlers.NewTemplateHandle(templateRepo)

	mux.Handle("POST /templates", authMiddleware(handler.CreateTemplateHandler()))
	mux.Handle("GET /templates", authMiddleware(handler.GetAllTemplatesHandler()))
	mux.Handle("GET /templates/{id}", authMiddleware(handler.GetTemplateHandler()))
	mux.Handle("PUT /templates/{id}", authMiddleware(handler.UpdateTemplateHandler()))
	mux.Handle("DELETE /templates/{id}", authMiddleware(handler.DeleteTemplateHandler()))

	mux.Handle("POST /templates/{id}/{type}/upload", authMiddleware(handler.UploadTemplateFileHandler()))
}
