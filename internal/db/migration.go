// File: /internal/db/migration.go

package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ApplyMigrations executa todas as migrations pendentes
func ApplyMigrations(db *sql.DB) error {
	migrationsDir := "migrations/" // Caminho relativo à raiz do projeto

	// Criar tabela de controle das migrations, se não existir
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS migrations (
		id SERIAL PRIMARY KEY,
		filename TEXT UNIQUE NOT NULL,
		applied_at TIMESTAMP DEFAULT NOW()
	)`)
	if err != nil {
		return fmt.Errorf("erro ao criar tabela de migrations: %v", err)
	}

	// Buscar migrations já aplicadas
	appliedMigrations := make(map[string]bool)
	rows, err := db.Query("SELECT filename FROM migrations")
	if err != nil {
		return fmt.Errorf("erro ao consultar migrations aplicadas: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			return err
		}
		appliedMigrations[filename] = true
	}

	// Listar arquivos de migrations na pasta
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("erro ao listar arquivos de migrations: %v", err)
	}

	// Ordenar por nome (garantir execução sequencial)
	var migrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}
	sort.Strings(migrations)

	// Aplicar apenas as migrations que ainda não foram executadas
	for _, migration := range migrations {
		if appliedMigrations[migration] {
			continue // Já foi aplicada, pular
		}

		filePath := filepath.Join(migrationsDir, migration)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("erro ao ler migration %s: %v", migration, err)
		}

		log.Printf("🚀 Aplicando migration: %s", migration)
		if _, err := db.Exec(string(content)); err != nil {
			return fmt.Errorf("erro ao executar migration %s: %v", migration, err)
		}

		// Registrar a migration aplicada
		_, err = db.Exec("INSERT INTO migrations (filename) VALUES ($1)", migration)
		if err != nil {
			return fmt.Errorf("erro ao registrar migration %s: %v", migration, err)
		}
	}

	log.Println("✅ Todas as migrations foram aplicadas com sucesso!")
	return nil
}
