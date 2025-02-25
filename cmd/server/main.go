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
	"github.com/jeancarlosdanese/go-marketing/internal/db/postgres"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/server/routes"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
)

func main() {
	// Carregar configurações do .env
	config.LoadConfig()

	// Agora sim podemos inicializar o logger
	logger.InitLogger()
	log := logger.GetLogger()

	// 🔥 Agora o banco é escolhido com base no .env (`DB_DRIVER=postgres`)
	dbConn, err := db.GetDatabase()
	if err != nil {
		logger.Fatal("Erro ao conectar ao banco de dados", err)
	}
	defer dbConn.Close() // 🔌 Fecha a conexão corretamente ao encerrar a aplicação

	// Criar repositórios
	otpRepo := postgres.NewAccountOTPRepository(dbConn)
	accountRepo := postgres.NewAccountRepository(dbConn)
	accountSettingsRepo := postgres.NewAccountSettingsRepository(dbConn)
	contactRepo := postgres.NewContactRepository(dbConn)
	templateRepo := postgres.NewTemplateRepository(dbConn)
	campaignRepo := postgres.NewCampaignRepository(dbConn)
	audienceRepo := postgres.NewCampaignAudienceRepository(dbConn)
	campaignSettingsRepo := postgres.NewCampaignSettingsRepository(dbConn)

	// 🔧 Inicializar serviços
	// Criar os serviços diretamente com os valores do ambiente
	sqsService, _ := service.NewSQSService(os.Getenv("SQS_EMAIL_URL"), os.Getenv("SQS_WHATSAPP_URL"))

	openAIService := service.NewOpenAIService()

	emailService := service.NewEmailService(accountSettingsRepo)

	whatsappService := service.NewWhatsAppService(
		os.Getenv("EVOLUTION_API_URL"),
		os.Getenv("EVOLUTION_API_KEY"),
		os.Getenv("EVOLUTION_INSTANCE"),
	)

	// 🚀 Iniciar os Workers
	workerService := service.NewWorkerService(
		sqsService,
		emailService,
		whatsappService,
		audienceRepo,
		contactRepo,
		campaignRepo,
		accountRepo,
		accountSettingsRepo,
		campaignSettingsRepo,
		openAIService,
	)
	workerService.Start(context.TODO())

	// Criar o servidor HTTP
	port := os.Getenv("APP_PORT")
	srv := &http.Server{
		Addr: fmt.Sprintf(":%s", port),
		Handler: routes.NewRouter(
			otpRepo, accountRepo,
			accountSettingsRepo,
			contactRepo,
			templateRepo,
			campaignRepo,
			audienceRepo,
			campaignSettingsRepo,
			openAIService,
			workerService,
		),
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
