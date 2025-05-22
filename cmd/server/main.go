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
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/server/routes"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/workers"
)

// Função para iniciar workers de e-mail e WhatsApp
func startWorker(ctx context.Context, worker workers.Worker, name string) {
	log := logger.GetLogger()
	go func() {
		log.Info(fmt.Sprintf("🚀 Iniciando %s...", name))
		worker.Start(ctx)
	}()
}

func main() {
	// Carregar configurações do .env
	config.LoadConfig()

	// Inicializar logger
	logger.InitLogger()
	log := logger.GetLogger()

	// Conectar ao banco de dados
	dbConn, err := db.GetDatabase()
	if err != nil {
		logger.Fatal("Erro ao conectar ao banco de dados", err)
	}
	defer dbConn.Close() // Fecha a conexão ao encerrar

	// Criar repositórios
	otpRepo := postgres.NewAccountOTPRepository(dbConn)
	accountRepo := postgres.NewAccountRepository(dbConn)
	accountSettingsRepo := postgres.NewAccountSettingsRepository(dbConn)
	contactRepo := postgres.NewContactRepository(dbConn)
	templateRepo := postgres.NewTemplateRepository(dbConn)
	campaignRepo := postgres.NewCampaignRepository(dbConn)
	audienceRepo := postgres.NewCampaignAudienceRepository(dbConn)
	campaignSettingsRepo := postgres.NewCampaignSettingsRepository(dbConn)
	contactImportRepo := postgres.NewContactImportRepository(dbConn)
	campaignMessageRepo := postgres.NewCampaignMessageRepository(dbConn)
	chatRepo := postgres.NewChatRepository(dbConn)
	chatContactRepo := postgres.NewChatContactRepository(dbConn)
	chatMessageRepo := postgres.NewChatMessageRepository(dbConn)

	// Inicializar serviços
	sqsService, _ := service.NewSQSService(os.Getenv("SQS_EMAIL_URL"), os.Getenv("SQS_WHATSAPP_URL"))
	openAIService := service.NewOpenAIService()
	campaignProcessor := service.NewCampaignProcessorService(sqsService, openAIService, audienceRepo)
	emailService := service.NewEmailService(openAIService)
	whatsappService := service.NewWhatsAppService(
		os.Getenv("EVOLUTION_API_URL"),
		os.Getenv("EVOLUTION_API_KEY"),
		os.Getenv("EVOLUTION_INSTANCE"),
	)

	// Criar contexto de controle para os workers
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Iniciar Workers de forma otimizada
	emailWorker := workers.NewEmailWorker(
		sqsService, emailService, audienceRepo, contactRepo, campaignRepo,
		accountRepo, accountSettingsRepo, campaignSettingsRepo, openAIService,
	)
	startWorker(ctx, emailWorker, "EmailWorker")

	whatsappWorker := workers.NewWhatsAppWorker(
		sqsService, whatsappService, audienceRepo, contactRepo, campaignRepo,
		accountRepo, accountSettingsRepo, campaignSettingsRepo, openAIService,
	)
	startWorker(ctx, whatsappWorker, "WhatsAppWorker")

	// Criar servidor HTTP com middleware CORS
	port := os.Getenv("APP_PORT")
	mux := http.NewServeMux()

	// 🔥 Aplica CORS a todas as rotas
	router := middleware.CORSMiddleware(routes.NewRouter(
		otpRepo, accountRepo, accountSettingsRepo, contactRepo,
		templateRepo, campaignRepo, audienceRepo, campaignSettingsRepo,
		openAIService, campaignProcessor, contactImportRepo,
		campaignMessageRepo, chatRepo, chatContactRepo, chatMessageRepo,
	))

	mux.Handle("/", router)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	// Capturar sinais do sistema
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Iniciar servidor em goroutine
	go func() {
		log.Info("🚀 Servidor rodando em http://localhost:" + port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Erro ao iniciar o servidor: " + err.Error())
		}
	}()

	// Bloqueia até receber um sinal de interrupção
	<-stop
	log.Info("⚠️  Sinal recebido! Encerrando servidor...")

	// Enviar sinal de cancelamento para os workers
	cancel()
	time.Sleep(2 * time.Second) // Dá tempo para os workers finalizarem

	// Criar contexto com timeout para desligamento do servidor
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Error("Erro ao desligar o servidor: " + err.Error())
	} else {
		log.Info("✅ Servidor encerrado com sucesso.")
	}
}
