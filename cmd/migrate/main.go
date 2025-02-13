// File: /cmd/migrate/main.go

package main

import (
	"fmt"
	"os"

	"github.com/jeancarlosdanese/go-marketing/config"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
)

func main() {
	fmt.Println("🚀 Iniciando aplicação de migrations...")

	// Carregar configurações do .env
	config.LoadConfig()

	// Inicializar o banco
	dbInstance, err := db.InitPostgresDB()
	if err != nil {
		fmt.Println("❌ Erro ao conectar ao banco:", err)
		os.Exit(1)
	}
	defer dbInstance.Close()

	// Aplicar migrations
	err = db.ApplyMigrations(dbInstance)
	if err != nil {
		fmt.Println("❌ Erro ao aplicar migrations:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Migrations aplicadas com sucesso!")
}
