// File: internal/logger/logger.go

package logger

import (
	"io"
	"log"
	"os"
	"runtime/debug"
	"sync"
)

// Níveis de log
const (
	DEBUG = "DEBUG"
	INFO  = "INFO"
	WARN  = "WARN"
	ERROR = "ERROR"
	FATAL = "FATAL"
)

// Logger estrutura o logger com múltiplos níveis
type Logger struct {
	debugLog *log.Logger
	infoLog  *log.Logger
	warnLog  *log.Logger
	errorLog *log.Logger
	fatalLog *log.Logger
	logFile  *os.File
}

var instance *Logger
var once sync.Once

// GetLogger retorna uma instância única do logger
func GetLogger() *Logger {
	once.Do(func() {
		// Criar diretório logs se não existir
		if _, err := os.Stat("logs"); os.IsNotExist(err) {
			os.Mkdir("logs", 0755)
		}

		// Criar ou abrir arquivo de log
		logFile, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("Erro ao abrir arquivo de log: %v", err)
		}

		multiWriter := io.MultiWriter(os.Stdout, logFile)

		instance = &Logger{
			debugLog: log.New(multiWriter, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
			infoLog:  log.New(multiWriter, "[INFO]  ", log.Ldate|log.Ltime|log.Lshortfile),
			warnLog:  log.New(multiWriter, "[WARN]  ", log.Ldate|log.Ltime|log.Lshortfile),
			errorLog: log.New(multiWriter, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
			fatalLog: log.New(multiWriter, "[FATAL] ", log.Ldate|log.Ltime|log.Lshortfile),
			logFile:  logFile,
		}
	})
	return instance
}

// Métodos para cada nível de log
func (l *Logger) Debug(msg string) { l.debugLog.Println(msg) }
func (l *Logger) Info(msg string)  { l.infoLog.Println(msg) }
func (l *Logger) Warn(msg string)  { l.warnLog.Println(msg) }

// Captura erro e stack trace
func (l *Logger) Error(msg string) {
	stack := debug.Stack()
	l.errorLog.Printf("%s\nStack Trace:\n%s", msg, string(stack))
}

func (l *Logger) Fatal(msg string) {
	stack := debug.Stack()
	l.fatalLog.Printf("%s\nStack Trace:\n%s", msg, string(stack))
	l.logFile.Close()
	os.Exit(1)
}
