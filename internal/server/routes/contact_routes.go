// File: /internal/server/routes/contact_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

// RegisterContactRoutes adiciona as rotas relacionadas a contatos
func RegisterContactRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, contactRepo db.ContactRepository, contactImportRepo db.ContactImportRepository, openAIService service.OpenAIService) {
	handler := handlers.NewContactHandle(contactRepo)

	importContactService := service.NewContactImportService(contactRepo, contactImportRepo, openAIService)

	// 📌 Importação de CSV
	importHandler := handlers.NewImportContactHandler(contactImportRepo, importContactService)

	// 🔒 Todas as rotas exigem autenticação
	mux.Handle("POST /contacts", authMiddleware(handler.CreateContactHandler()))        // Criar contato
	mux.Handle("GET /contacts", authMiddleware(handler.GetPaginatedContactsHandler()))  // Listar contatos da conta autenticada
	mux.Handle("GET /contacts/{id}", authMiddleware(handler.GetContactHandler()))       // Buscar contato por ID
	mux.Handle("PUT /contacts/{id}", authMiddleware(handler.UpdateContactHandler()))    // Atualizar contato
	mux.Handle("DELETE /contacts/{id}", authMiddleware(handler.DeleteContactHandler())) // Deletar contato

	// 📌 Importação de CSV
	// mux.Handle("POST /contacts/import", authMiddleware(importHandler.UploadCSVHandler())) // Importar CSV
	mux.Handle("POST /contacts/import", authMiddleware(importHandler.UploadHandler()))                  // Importar CSV
	mux.Handle("GET /contacts/imports", authMiddleware(importHandler.GetImportsHandler()))              // Listar importações
	mux.Handle("POST /contacts/imports/{id}/start", authMiddleware(importHandler.StartImportHandler())) // Iniciar processamento
	mux.Handle("GET /contacts/imports/{id}", authMiddleware(importHandler.GetImportByIDHandler()))      // Listar importações
	mux.Handle("PUT /contacts/imports/{id}", authMiddleware(importHandler.UpdateImportConfigHandler())) // Atualizar configuração de importação
	mux.Handle("DELETE /contacts/imports/{id}", authMiddleware(importHandler.RemoveImportHandler()))    // Remover importação
}
