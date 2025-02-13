// File: /internal/observability/otel.go

package observability

import (
	"context"
	"fmt"
	"log"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracerProvider *sdktrace.TracerProvider
	once           sync.Once
)

// InitTracer inicializa o OpenTelemetry e exporta para stdout (terminal)
func InitTracer() func() {
	fmt.Println("🔧 Iniciando OpenTelemetry...")

	once.Do(func() {
		// Criar um exportador que imprime no terminal
		exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			log.Fatalf("❌ Erro ao criar exportador OpenTelemetry: %v", err)
		}

		// Criar provedor de traces e definir corretamente os atributos do serviço
		tracerProvider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter), // Exportador para stdout
			sdktrace.WithResource(resource.NewSchemaless(
				attribute.String("service.name", "go-marketing"), // Nome do serviço corrigido
				attribute.String("service.version", "1.0.0"),
			)),
		)

		// Definir provedor global de traces
		otel.SetTracerProvider(tracerProvider)

		fmt.Println("✅ OpenTelemetry inicializado com sucesso!")
	})

	// Retorna uma função para encerrar corretamente o tracer ao finalizar o programa
	return func() {
		if err := tracerProvider.Shutdown(context.Background()); err != nil {
			log.Fatalf("❌ Erro ao encerrar o tracer: %v", err)
		}
	}
}

// GetTracer retorna um tracer para instrumentação
func GetTracer(name string) trace.Tracer {
	if tracerProvider == nil {
		log.Fatalf("❌ Erro: Tracer Provider não foi inicializado antes de chamar GetTracer()")
	}
	fmt.Println("📌 Obtendo tracer:", name)
	return tracerProvider.Tracer(name)
}
