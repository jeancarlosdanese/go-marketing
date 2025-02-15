// File: /internal/server/routes/routes.go

package routes

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
)

// NewRouter cria e retorna um roteador HTTP configurado.
func NewRouter(otpRepo db.AccountOTPRepository, accountRepo db.AccountRepository, accountSettingsRepo db.AccountSettingsRepository) *http.ServeMux {
	mux := http.NewServeMux()

	// 🔥 Criar middlewares
	authMiddleware := middleware.AuthMiddleware(accountRepo)

	// 🔥 Registrar rotas principais
	RegisterAuthRoutes(mux, otpRepo)
	RegisterAccountRoutes(mux, authMiddleware, accountRepo)
	RegisterAccountSettingsRoutes(mux, authMiddleware, accountSettingsRepo)

	// 🔥 Rota de Health Check
	mux.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}))

	return mux
}
