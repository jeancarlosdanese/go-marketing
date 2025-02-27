// File: internal/server/routes/auth_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterAuthRoutes adiciona as rotas relacionadas à autenticação
func RegisterAuthRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.HandlerFunc, otpRepo db.AccountOTPRepository) {
	handler := handlers.NewAuthHandle(otpRepo)

	mux.HandleFunc("POST /auth/request-otp", handler.RequestAuthHandle())
	mux.HandleFunc("POST /auth/verify-otp", handler.VerifyAuthHandle())

	// Adiciona rota para obter informações do usuário autenticado
	mux.HandleFunc("GET /auth/me", authMiddleware(handler.MeHandler()))
}
