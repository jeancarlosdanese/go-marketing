// File: /config/config.go

package config

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Config armazena as variáveis de configuração
type Config struct {
	mu      sync.RWMutex
	envFile string
}

// LoadConfig carrega as variáveis do .env e inicia o monitoramento
func LoadConfig() {
	log.Println("🔥 Carregando configurações do .env")

	projectRoot, err := os.Getwd()
	if err != nil {
		log.Fatalf("❌ Erro ao obter diretório de trabalho: %v", err)
	}

	envPath := filepath.Join(projectRoot, ".env")

	cfg := &Config{envFile: envPath}
	cfg.loadFromFile()
}

// loadFromFile lê e aplica as variáveis do .env
func (c *Config) loadFromFile() {
	c.mu.Lock()
	defer c.mu.Unlock()

	file, err := os.Open(c.envFile)
	if err != nil {
		log.Fatalf("❌ Erro ao abrir o arquivo .env: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`) // 🔥 Remove aspas extras, se houver
		os.Setenv(key, value)
	}
}

func GetEnvVar(key string) string {
	return os.Getenv(key)
}
