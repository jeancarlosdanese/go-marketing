// File: internal/middleware/auth_middleware.go

package middleware

import (
	"net/http"
	"strings"

	"github.com/jeancarlosdanese/go-marketing/internal/auth"
)

func AuthMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		tokenStr := tokenParts[1]
		_, err := auth.ValidateJWT(tokenStr)
		if err != nil {
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	}
}
