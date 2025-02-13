// File: /internal/server/router.go

package server

import (
	"fmt"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/observability"
)

// NewRouter cria e retorna um roteador HTTP
func NewRouter() http.Handler {
	mux := http.NewServeMux()
	tracer := observability.GetTracer("http-server")

	// Rota de saúde com tracing
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("🔎 Recebida requisição em /health")

		// Criar span com contexto correto
		_, span := tracer.Start(r.Context(), "health-check")
		defer span.End()

		traceID := span.SpanContext().TraceID()
		fmt.Println("✅ OpenTelemetry Trace ID:", traceID.String())

		fmt.Fprintln(w, "OK")
		span.AddEvent("Health check concluído")
	})

	return mux
}
