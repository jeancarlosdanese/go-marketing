// File: /internal/db/db_postgres.go

package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	postgresInstance *sql.DB
	oncePostgres     sync.Once
)

// InitPostgresDB inicializa a conexão com o PostgreSQL
func InitPostgresDB() (*sql.DB, error) {
	var err error
	oncePostgres.Do(func() {
		dsn := getPostgresDSN() // 🔥 Obtém a string de conexão

		postgresInstance, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Fatalf("❌ Erro ao conectar ao PostgreSQL: %v", err)
		}

		// 🔥 Configuração do Pool de Conexões
		postgresInstance.SetMaxOpenConns(25)
		postgresInstance.SetMaxIdleConns(5)
		postgresInstance.SetConnMaxLifetime(0)

		log.Println("✅ Conexão com PostgreSQL inicializada com sucesso!")
	})

	return postgresInstance, err
}

// Função privada para obter a string de conexão
func getPostgresDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
}
