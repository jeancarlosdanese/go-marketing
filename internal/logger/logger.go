// File: internal/logger/logger.go

// Package logger é responsável por criar uma instância única do logger
// para ser utilizado em todo o sistema.
// O logger é baseado no pacote log/slog que é uma extensão do pacote log padrão
// do Go, porém com suporte a logs estruturados em JSON.
// O logger é inicializado apenas uma vez e é retornado sempre a mesma instância
// para ser utilizado em todo o sistema.
// O logger é inicializado com o nível de log INFO, ou seja, apenas logs de nível
// INFO ou superior serão exibidos.
// Para utilizar o logger, basta importar o pacote e chamar a função GetLogger().
// Exemplo:
//
//	logger := logger.GetLogger()
//	logger.Info("Mensagem de log")
//	logger.Warn("Mensagem de log")
//	logger.Error("Mensagem de log")
//	logger.Debug("Mensagem de log")
//	logger.WithField("campo", "valor").Info("Mensagem de log")
//	logger.WithFields(slog.Fields{"campo1": "valor1", "campo2": "valor2"}).Info("Mensagem de log")
//	logger.WithError(err).Error("Mensagem de log")
//	logger.WithError(err).WithField("campo", "valor").Error("Mensagem de log")
//	logger.WithError(err).WithFields(slog.Fields{"campo1": "valor1", "campo2": "valor2"}).Error("Mensagem de log")
//	logger.WithFields(slog.Fields{"campo1": "valor1", "campo2": "valor2"}).WithError(err).Error("Mensagem de log")
//	logger.WithFields(slog.Fields{"campo1": "valor1", "campo2": "valor2"}).WithField("campo", "valor").WithError(err).Error("Mensagem de log")
//	logger.WithFields(slog.Fields{"campo1": "valor1", "campo2": "valor2"}).WithField("campo", "valor").WithError(err).Info("Mensagem de log")
//	logger.WithFields(slog.Fields{"campo1": "valor1", "campo2": "valor2"}).WithField("campo", "valor").WithError(err).Debug("Mensagem de log")

package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

// Logger é a estrutura do sistema de logs
var log *slog.Logger
var once sync.Once

// GetLogger retorna uma instância única do logger
func GetLogger() *slog.Logger {
	once.Do(func() {
		// Verificar se `APP_DEBUG` está ativado
		logLevel := slog.LevelInfo // Padrão (INFO, WARN, ERROR, FATAL)
		if strings.Contains(os.Getenv("APP_MODE"), "dev") {
			logLevel = slog.LevelDebug // Ativar logs de DEBUG se APP_DEBUG=true
		}

		// Criar um handler apenas para o console
		consoleHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		})

		// Criar logger
		log = slog.New(consoleHandler)
	})
	return log
}
