// File: internal/middleware/auth_middleware.go

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/auth"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// contextKeyAccount é uma chave única para armazenar a conta no contexto
type contextKeyAccount struct{}

var AuthAccountKey contextKeyAccount = struct{}{}

// AuthMiddleware recebe o repositório e adiciona o `account_id` + perfil no contexto
func AuthMiddleware(accountRepo db.AccountRepository) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Token não fornecido", http.StatusUnauthorized)
				return
			}

			accountID, err := auth.GetAccountIDFromToken(r)
			if err != nil {
				http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
				return
			}

			// Buscar a conta autenticada no banco (reaproveitando a conexão!)
			uuidAccountID, err := uuid.Parse(accountID)
			if err != nil {
				http.Error(w, "ID da conta inválido", http.StatusInternalServerError)
				return
			}

			account, err := accountRepo.GetByID(uuidAccountID)
			if err != nil {
				http.Error(w, "Conta não encontrada", http.StatusUnauthorized)
				return
			}

			// Adiciona a conta autenticada no contexto
			ctx := context.WithValue(r.Context(), AuthAccountKey, account)

			// Passa a requisição para o próximo handler
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// 📌 **Helper para recuperar a conta autenticada**
func GetAuthenticatedAccount(ctx context.Context) (*models.Account, bool) {
	account, ok := ctx.Value(AuthAccountKey).(*models.Account)
	return account, ok
}
