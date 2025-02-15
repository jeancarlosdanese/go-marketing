// File: /internal/server/routes/contact_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterContactRoutes adiciona as rotas relacionadas a contatos
func RegisterContactRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, contactRepo db.ContactRepository) {
	handler := handlers.NewContactHandle(contactRepo)

	// 🔒 Todas as rotas exigem autenticação
	mux.Handle("POST /contacts", authMiddleware(handler.CreateContactHandler()))        // Criar contato
	mux.Handle("GET /contacts", authMiddleware(handler.GetAllContactsHandler()))        // Listar contatos da conta autenticada
	mux.Handle("GET /contacts/{id}", authMiddleware(handler.GetContactHandler()))       // Buscar contato por ID
	mux.Handle("PUT /contacts/{id}", authMiddleware(handler.UpdateContactHandler()))    // Atualizar contato
	mux.Handle("DELETE /contacts/{id}", authMiddleware(handler.DeleteContactHandler())) // Deletar contato
}
