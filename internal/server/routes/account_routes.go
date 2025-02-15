// File: /internal/server/routes/account_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterAccountRoutes adiciona as rotas relacionadas a contas
func RegisterAccountRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, accountRepo db.AccountRepository) {
	handler := handlers.NewAccountHandle(accountRepo)

	mux.Handle("POST /accounts", http.HandlerFunc(handler.CreateAccountHandler()))
	mux.Handle("GET /accounts", authMiddleware(http.HandlerFunc(handler.GetAllAccountsHandler())))
	mux.Handle("GET /accounts/{id}", authMiddleware(http.HandlerFunc(handler.GetAccountHandler())))
	mux.Handle("PUT /accounts/{id}", authMiddleware(http.HandlerFunc(handler.UpdateAccountHandler())))
	mux.Handle("DELETE /accounts/{id}", authMiddleware(http.HandlerFunc(handler.DeleteAccountHandler())))
}
