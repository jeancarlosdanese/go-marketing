// File: cmd/test_crud/main.go

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jeancarlosdanese/go-marketing/config"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/db/postgres"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

func main() {
	fmt.Println("🚀 Iniciando testes do CRUD de contas...")

	// Carregar configurações do .env
	config.LoadConfig()

	// Inicializar banco
	dbInstance, err := db.InitPostgresDB()
	if err != nil {
		fmt.Println("❌ Erro ao conectar ao banco:", err)
		os.Exit(1)
	}
	defer dbInstance.Close()

	// Criar repositório de contas
	accountRepo := postgres.NewAccountRepository(dbInstance)

	// Criar nova conta
	newAccount := &models.Account{
		Name:  "Teste User",
		Email: "teste@example.com",
	}
	account, err := accountRepo.Create(context.TODO(), newAccount)
	if err != nil {
		fmt.Println("❌ Erro ao criar conta:", err)
	}

	fmt.Printf("✅ Conta criada com sucesso: %+v\n", account)

	// Buscar conta pelo ID
	foundAccount, err := accountRepo.GetByID(context.TODO(), newAccount.ID)
	if err != nil {
		fmt.Println("❌ Erro ao buscar conta:", err)
	} else {
		fmt.Printf("🔍 Conta encontrada: %+v\n", foundAccount)
	}
}
