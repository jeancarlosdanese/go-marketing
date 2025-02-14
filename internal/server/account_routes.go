// File: /internal/server/account_routes.go

package server

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterAccountRoutes adiciona as rotas relacionadas a contas
func RegisterAccountRoutes(mux *http.ServeMux, accountRepo db.AccountRepository) {
	handler := handlers.NewAccountHandle(accountRepo)

	mux.HandleFunc("POST /accounts", handler.CreateAccountHandler())
	mux.HandleFunc("GET /accounts", middleware.AuthMiddleware(http.HandlerFunc(handler.GetAllAccountsHandler())))
	mux.HandleFunc("GET /accounts/", handler.GetAccountHandler())
	mux.HandleFunc("PUT /accounts/", handler.UpdateAccountHandler())
	mux.HandleFunc("DELETE /accounts/", handler.DeleteAccountHandler())
}
