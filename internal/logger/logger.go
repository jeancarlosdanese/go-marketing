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
)

// Ícones associados a cada nível de log
var logIcons = map[slog.Leveler]string{
	slog.LevelDebug: "🐛 DEBUG:",
	slog.LevelInfo:  "ℹ️ INFO:",
	slog.LevelWarn:  "⚠️ WARN:",
	slog.LevelError: "❌ ERROR:",
}

var logInstance *slog.Logger

// InitLogger deve ser chamado após carregar as configurações do ambiente
func InitLogger() {
	// Define o nível de log com base no APP_MODE
	level := slog.LevelInfo
	if strings.Contains(os.Getenv("APP_MODE"), "dev") {
		level = slog.LevelDebug
	}

	// Definir saída JSON ou texto legível com base no ambiente
	var handler slog.Handler
	if os.Getenv("LOG_FORMAT") == "json" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	logInstance = slog.New(handler)
}

// GetLogger retorna o logger configurado
func GetLogger() *slog.Logger {
	if logInstance == nil {
		InitLogger() // Garante que o logger está inicializado
	}
	return logInstance
}

// Debug loga mensagens de debug (aparecem apenas se APP_MODE for "dev")
func Debug(msg string, args ...any) {
	logInstance.Debug(logIcons[slog.LevelDebug]+" "+msg, args...)
}

// Info loga mensagens informativas
func Info(msg string, args ...any) {
	logInstance.Info(logIcons[slog.LevelInfo]+" "+msg, args...)
}

// Warn loga avisos
func Warn(msg string, args ...any) {
	logInstance.Warn(logIcons[slog.LevelWarn]+" "+msg, args...)
}

// Error loga erros
func Error(msg string, err error, args ...any) {
	logInstance.Error(logIcons[slog.LevelError]+" "+msg, append(args, "error", err.Error())...)
}

// Fatal loga erro fatal e encerra o programa
func Fatal(msg string, err error, args ...any) {
	logInstance.Error("💀 FATAL ERROR: "+msg, append(args, "error", err.Error())...)
	os.Exit(1) // Encerra o programa
}

// Panic loga erro crítico e faz panic
func Panic(msg string, err error, args ...any) {
	logInstance.Error("🔥 PANIC ERROR: "+msg, append(args, "error", err.Error())...)
	panic(msg) // Dispara um panic
}
