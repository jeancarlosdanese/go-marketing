// File: /config/config.go

package config

import (
	"bufio"
	"log"
	"os"
	"strings"
	"sync"
)

// Config armazena as variáveis de configuração
type Config struct {
	mu                 sync.RWMutex
	DBDriver           string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	OpenAIAPIKey       string
	EvolutionAPIURL    string
	EvolutionAPIKey    string
	EvolutionInstance  string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	MailFrom           string
	MailAdminTo        string
	XAPIKey            string
	LimitDay           string
	envFile            string
}

// AppConfig é a instância global das configurações
var AppConfig *Config

// LoadConfig carrega as variáveis do .env e inicia o monitoramento
func LoadConfig(envFile string) *Config {
	cfg := &Config{envFile: envFile}
	cfg.loadFromFile()

	AppConfig = cfg
	return cfg
}

// loadFromFile lê e aplica as variáveis do .env
func (c *Config) loadFromFile() {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.envFile)
	if err != nil {
		log.Fatalf("Erro ao abrir o arquivo .env: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") || len(strings.TrimSpace(line)) == 0 {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "DB_DRIVER":
			c.DBDriver = value
		case "DB_HOST":
			c.DBHost = value
		case "DB_PORT":
			c.DBPort = value
		case "DB_USER":
			c.DBUser = value
		case "DB_PASSWORD":
			c.DBPassword = value
		case "DB_NAME":
			c.DBName = value
		case "OPENAI_API_KEY":
			c.OpenAIAPIKey = value
		case "EVOLUTION_API_URL":
			c.EvolutionAPIURL = value
		case "EVOLUTION_API_KEY":
			c.EvolutionAPIKey = value
		case "EVOLUTION_INSTANCE":
			c.EvolutionInstance = value
		case "AWS_ACCESS_KEY_ID":
			c.AWSAccessKeyID = value
		case "AWS_SECRET_ACCESS_KEY":
			c.AWSSecretAccessKey = value
		case "AWS_REGION":
			c.AWSRegion = value
		case "MAIL_FROM":
			c.MailFrom = value
		case "MAIL_ADMIN_TO":
			c.MailAdminTo = value
		case "X_API_KEY":
			c.XAPIKey = value
		case "LIMIT_DAY":
			c.LimitDay = value
		}
	}
}
