// File: /internal/db/db.go

package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

// Database representa uma interface genérica para bancos de dados
type Database interface {
	Init() (*sql.DB, error)
}

// GetDatabase retorna a implementação do banco com base na configuração
func GetDatabase() (*sql.DB, error) {
	dbType := os.Getenv("DB_DRIVER") // 🔥 Definir no .env: postgres, mysql, sqlite

	switch dbType {
	case "postgres":
		return InitPostgresDB()
	case "mysql":
		return nil, fmt.Errorf("banco de dados mysql não implementado")
	default:
		log.Fatalf("❌ Banco de dados '%s' não suportado!", dbType)
		return nil, fmt.Errorf("banco de dados não suportado: %s", dbType)
	}
}
