// File: internal/server/routes/auth_routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/server/handlers"
)

// RegisterAuthRoutes adiciona as rotas relacionadas à autenticação
func RegisterAuthRoutes(mux *http.ServeMux, otpRepo db.AccountOTPRepository) {
	handler := handlers.NewAuthHandle(otpRepo)

	mux.HandleFunc("POST /auth/request-otp", handler.RequestAuthHandle())
	mux.HandleFunc("POST /auth/verify-otp", handler.VerifyAuthHandle())
}
