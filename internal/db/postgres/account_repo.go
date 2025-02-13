// File: /internal/db/postgres/account_repo.go

package postgres

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// AccountRepository implementa db.AccountRepository para PostgreSQL
type AccountRepository struct {
	db *sql.DB
}

// NewAccountRepository cria um novo repositório recebendo a conexão como argumento
func NewAccountRepository(dbConn *sql.DB) db.AccountRepository {
	if dbConn == nil {
		panic("❌ Banco de dados não inicializado! Chame InitPostgresDB() primeiro.")
	}
	return &AccountRepository{db: dbConn}
}

// Create insere um novo registro na tabela accounts
func (r *AccountRepository) Create(account *models.Account) (*models.Account, error) {
	account.ID = uuid.New() // 🔥 Gerar UUID antes de salvar

	query := "INSERT INTO accounts (id, name, email) VALUES ($1, $2, $3) RETURNING id"
	err := r.db.QueryRow(query, account.ID, account.Name, account.Email).Scan(&account.ID)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// GetByID busca uma conta pelo ID (usando UUID)
func (r *AccountRepository) GetByID(id uuid.UUID) (*models.Account, error) {
	query := "SELECT id, name, email FROM accounts WHERE id = $1"
	row := r.db.QueryRow(query, id)

	account := &models.Account{}
	err := row.Scan(&account.ID, &account.Name, &account.Email)
	if err != nil {
		return nil, err
	}
	return account, nil
}
