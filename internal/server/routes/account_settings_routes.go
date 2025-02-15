// File: /internal/server/routes/account_settings_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterAccountRoutes adiciona as rotas relacionadas a contas
func RegisterAccountSettingsRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, accountSettingsRepo db.AccountSettingsRepository) {

	handler := handlers.NewAccountSettingsHandle(accountSettingsRepo)

	mux.Handle("POST /account-settings", authMiddleware(handler.CreateAccountSettingsHandler()))

	// 🔒 Apenas o dono da conta pode acessar essa rota
	mux.Handle("GET /account-settings", authMiddleware(handler.GetAccountSettingsHandler()))
	mux.Handle("DELETE /account-settings", authMiddleware(handler.DeleteAccountSettingsHandler()))

	// 🔒 Apenas administradores podem acessar essa rota
	mux.Handle("GET /account-settings/{account_id}", authMiddleware(handler.GetAccountSettingsHandler()))
	mux.Handle("DELETE /account-settings/{account_id}", authMiddleware(handler.DeleteAccountSettingsHandler()))

	mux.Handle("PUT /account-settings", authMiddleware(handler.UpdateAccountSettingsHandler()))
}
