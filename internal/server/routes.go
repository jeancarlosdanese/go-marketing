package server

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
)

// NewRouter cria e retorna um roteador HTTP configurado.
func NewRouter(accountRepo db.AccountRepository, otpRepo db.AccountOTPRepository) *http.ServeMux {
	mux := http.NewServeMux()

	// 🔥 Criar middlewares
	authMiddleware := middleware.AuthMiddleware(accountRepo)

	// 🔥 Registrar rotas principais
	RegisterAccountRoutes(mux, authMiddleware, accountRepo)
	RegisterAuthRoutes(mux, otpRepo)

	// 🔥 Rota de Health Check
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))

	return mux
}
