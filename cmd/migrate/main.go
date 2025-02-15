// File: /cmd/migrate/main.go

package main

import (
	"fmt"
	"os"

	"github.com/jeancarlosdanese/go-marketing/config"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

func main() {
	fmt.Println("🚀 Iniciando aplicação de migrations...")

	// Carregar configurações do .env
	config.LoadConfig()

	// Agora sim podemos inicializar o logger
	logger.InitLogger()

	// 🔥 Agora o banco é escolhido com base no .env (`DB_DRIVER=postgres`)
	dbConn, err := db.GetDatabase()
	if err != nil {
		logger.Fatal("Erro ao conectar ao banco de dados", err)
	}
	defer dbConn.Close() // 🔌 Fecha a conexão corretamente ao encerrar a aplicação

	// Aplicar migrations
	err = db.ApplyMigrations(dbConn)
	if err != nil {
		fmt.Println("❌ Erro ao aplicar migrations:", err)
		os.Exit(1)
	}

	fmt.Println("✅ Migrations aplicadas com sucesso!")
}
