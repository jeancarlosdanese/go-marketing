// File: /internal/server/router.go

package server

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
)

// NewRouter cria um roteador HTTP modular
func NewRouter(accountRepo db.AccountRepository, otpRepo db.AccountOTPRepository) *http.ServeMux {
	mux := http.NewServeMux()

	// 🔥 Registrar rotas específicas para cada módulo
	RegisterAccountRoutes(mux, accountRepo)
	RegisterAuthRoutes(mux, otpRepo)

	// Health Check
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	return mux
}
