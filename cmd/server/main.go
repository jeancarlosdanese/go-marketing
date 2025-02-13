// File: /cmd/server/main.go

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeancarlosdanese/go-marketing/config"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/observability"
	"github.com/jeancarlosdanese/go-marketing/internal/server"
)

func main() {
	// Inicializar logger
	log := logger.GetLogger()

	// Inicializar OpenTelemetry ANTES de carregar o router
	cleanup := observability.InitTracer()
	defer cleanup()

	// Carregar configurações do .env
	config.LoadConfig(".env")
	log.Info("Configurações carregadas.")

	// Criar o servidor HTTP
	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.NewRouter(),
	}

	// Canal para capturar sinais do sistema
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar o servidor em uma goroutine
	go func() {
		log.Info("🚀 Servidor rodando em http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Erro ao iniciar o servidor: " + err.Error())
		}
	}()

	// Bloqueia até receber um sinal de interrupção
	<-stop
	log.Info("⚠️  Sinal recebido! Encerrando servidor...")

	// Criar um contexto com timeout para shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("Erro ao desligar o servidor: " + err.Error())
	} else {
		log.Info("✅ Servidor encerrado com sucesso.")
	}
}
