// internal/middleware/auth_middleware.go

package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/auth"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// contextKeyAccount é uma chave única para armazenar a conta no contexto.
type contextKeyAccount struct{}

// AuthAccountKey é a key usada para buscar a conta do contexto.
var AuthAccountKey contextKeyAccount = struct{}{}

// InternalAPIKey é a chave de API interna usada para autenticação.
const InternalAPIKeyHeader = "X-API-Key"

// InternalAPIKey é carregado do .env
var InternalAPIKey = os.Getenv("INTERNAL_API_KEY")

// AuthMiddleware recebe o repositório e injeta a `Account` autenticada no contexto.
func AuthMiddleware(accountRepo db.AccountRepository) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.SendError(w, http.StatusUnauthorized, "Token não fornecido")
				return
			}

			accountIDStr, err := auth.GetAccountIDFromToken(r)
			if err != nil {
				utils.SendError(w, http.StatusUnauthorized, "Token inválido ou expirado")
				return
			}

			uuidAccountID, err := uuid.Parse(accountIDStr)
			if err != nil {
				utils.SendError(w, http.StatusBadRequest, "ID da conta inválido no token")
				return
			}

			account, err := accountRepo.GetByID(context.Background(), uuidAccountID)
			if err != nil || account == nil {
				utils.SendError(w, http.StatusUnauthorized, "Conta não encontrada ou inexistente")
				return
			}

			// Adiciona a conta autenticada no contexto.
			ctx := context.WithValue(r.Context(), AuthAccountKey, account)

			// Passa a requisição para o próximo handler.
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// GetAuthenticatedAccount recupera a conta autenticada do contexto.
func GetAuthenticatedAccount(ctx context.Context) (*models.Account, bool) {
	account, ok := ctx.Value(AuthAccountKey).(*models.Account)
	return account, ok
}

// GetAuthAccountOrFail retorna a conta autenticada ou responde com erro (500) se não existir.
func GetAuthAccountOrFail(ctx context.Context, w http.ResponseWriter, log *slog.Logger) *models.Account {
	authAccount, ok := GetAuthenticatedAccount(ctx)
	if !ok || authAccount == nil {
		log.Error("Conta não encontrada no contexto")
		utils.SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
		return nil
	}
	return authAccount
}

// IsAdminOrOwner ajuda a reduzir duplicação de lógica: se não for admin, verifica se é o dono do recurso.
func IsAdminOrOwner(account *models.Account, ownerID uuid.UUID) bool {
	// A struct Account possui IsAdmin() ou logicamente definimos que ID == "000000..." é admin?
	// Caso exista a função `IsAdmin()`, use-a. Exemplo:
	// if account.IsAdmin() { return true }

	// Se você não tiver a IsAdmin(), mas sim um ID fixo, pode checar:
	// adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	// if account.ID == adminID {
	//     return true
	// }

	// Se não for admin, retorna true apenas se for o dono do recurso.
	// Exemplo:
	return account.IsAdmin() || (account.ID == ownerID)
}

// InternalAPIKeyMiddleware valida chamadas internas entre serviços (ex: Node.js -> Go)
func InternalAPIKeyMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get(InternalAPIKeyHeader)
		if key == "" || key != InternalAPIKey {
			utils.SendError(w, http.StatusUnauthorized, "Chave de API interna inválida")
			return
		}

		next.ServeHTTP(w, r)
	}
}
