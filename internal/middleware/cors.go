// File: /internal/middleware/cors.go

package middleware

import (
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

// CORSMiddleware adiciona os cabeçalhos CORS corretamente
func CORSMiddleware(next http.Handler) http.Handler {
	log := logger.GetLogger()

	log.Debug("Adicionando middleware CORS")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Middleware CORS", "method", r.Method, "path", r.URL.Path)

		// 🔥 Permitir chamadas do frontend no localhost
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")

		// 🔥 Permitir métodos usados pelo frontend
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// 🔥 Permitir headers necessários
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// 🔥 Permitir credenciais (se necessário)
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// ✅ Responder diretamente às requisições OPTIONS (preflight)
		if r.Method == http.MethodOptions {
			log.Debug("CORS: Respondendo OPTIONS com 204 No Content", "path", r.URL.Path)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 🔥 Passa para o próximo handler se não for OPTIONS
		next.ServeHTTP(w, r)
	})
}
