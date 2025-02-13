// File: /cmd/server/main.go

package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jeancarlosdanese/go-marketing/config"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/server"
)

func main() {
	// Carregar configurações do .env
	config.LoadConfig()

	// Inicializar logger
	log := logger.GetLogger()
	log.Info("Configurações carregadas.")

	// 🔥 Agora o banco é escolhido com base no .env (`DB_DRIVER=postgres`)
	dbInstance, err := db.GetDatabase()
	if err != nil {
		fmt.Println("Erro ao conectar ao banco:", err)
		os.Exit(1)
	}
	defer dbInstance.Close() // 🔌 Fecha a conexão corretamente ao encerrar a aplicação

	// 🔥 Criar repositório usando a conexão passada como argumento
	// repo := postgres.NewAccountRepository(dbInstance)

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
