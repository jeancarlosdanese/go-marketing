// File: /internal/server/router.go

package server

import (
	"fmt"
	"net/http"
)

// NewRouter cria e retorna um roteador HTTP
func NewRouter() http.Handler {
	mux := http.NewServeMux()

	// Rota de saúde com tracing
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "OK")
	})

	return mux
}
