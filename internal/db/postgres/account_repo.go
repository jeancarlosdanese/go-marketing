// File: /internal/db/postgres/account_repo.go

package postgres

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// AccountRepoPostgres implementa db.AccountRepository para PostgreSQL
type AccountRepoPostgres struct {
	log *slog.Logger
	db  *sql.DB
}

// NewAccountRepository cria um novo repositório recebendo a conexão como argumento
func NewAccountRepository(db *sql.DB) db.AccountRepository {
	log := logger.GetLogger()

	if db == nil {
		panic("❌ Banco de dados não inicializado! Chame InitPostgresDB() primeiro.")
	}

	return &AccountRepoPostgres{log: log, db: db}
}

// Create insere um novo registro na tabela accounts
func (r *AccountRepoPostgres) Create(account *models.Account) (*models.Account, error) {
	r.log.Debug("Criando nova conta", "name", account.Name, "email", account.Email)

	account.ID = uuid.New() // 🔥 Gerar UUID antes de salvar

	query := "INSERT INTO accounts (id, name, email, whatsapp) VALUES ($1, $2, $3, $4) RETURNING id"

	err := r.db.QueryRow(query, account.ID, account.Name, account.Email, account.WhatsApp).Scan(&account.ID)
	if err != nil {
		r.log.Error("Erro ao inserir nova conta", "error", err)
		return nil, err
	}

	return account, nil
}

// GetByID busca uma conta pelo ID (usando UUID)
func (r *AccountRepoPostgres) GetByID(id uuid.UUID) (*models.Account, error) {
	r.log.Debug("Buscando conta por ID", "id", id)

	query := "SELECT id, name, email, whatsapp FROM accounts WHERE id = $1"
	row := r.db.QueryRow(query, id)

	account := &models.Account{}
	err := row.Scan(&account.ID, &account.Name, &account.Email, &account.WhatsApp)
	if err != nil {
		return nil, err
	}
	return account, nil
}

// GetAll busca todas as contas cadastradas
func (r *AccountRepoPostgres) GetAll() ([]*models.Account, error) {
	r.log.Debug("Buscando todas as contas")

	query := "SELECT id, name, email, whatsapp FROM accounts ORDER BY name ASC"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*models.Account
	for rows.Next() {
		account := &models.Account{}
		if err := rows.Scan(&account.ID, &account.Name, &account.Email, &account.WhatsApp); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, nil
}

// UpdateByID atualiza uma conta a partir de um JSON
func (r *AccountRepoPostgres) UpdateByID(id uuid.UUID, jsonData []byte) (*models.Account, error) {
	r.log.Debug("Atualizando conta por ID", "id", id)

	// Decodificar o JSON para obter os campos a serem atualizados
	var updatedData map[string]interface{}
	if err := json.Unmarshal(jsonData, &updatedData); err != nil {
		r.log.Error("Erro ao decodificar JSON", "error", err)
		return nil, err
	}

	r.log.Debug("Atualizando conta", "id", id, "data", updatedData)

	// Construir a query dinamicamente
	query := "UPDATE accounts SET "
	values := []interface{}{id}
	i := 2

	for key, value := range updatedData {
		if key != "id" { // Nunca permitir alteração do ID
			query += key + " = $" + strconv.Itoa(i) + ", "
			values = append(values, value)
			i++
		}
	}

	// Remover a última vírgula e adicionar a cláusula WHERE
	query = query[:len(query)-2] + " WHERE id = $1 RETURNING id, name, email, whatsapp"

	// Executar a query
	account := &models.Account{}
	err := r.db.QueryRow(query, values...).Scan(&account.ID, &account.Name, &account.Email, &account.WhatsApp)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// DeleteByID remove uma conta pelo ID
func (r *AccountRepoPostgres) DeleteByID(id uuid.UUID) (uuid.UUID, error) {
	r.log.Debug("Deletando conta", "id", id)

	query := "DELETE FROM accounts WHERE id = $1 RETURNING id"
	var deletedID uuid.UUID

	err := r.db.QueryRow(query, id).Scan(&deletedID)
	if err != nil {
		if err == sql.ErrNoRows {
			return uuid.Nil, errors.New("conta não encontrada")
		}
		return uuid.Nil, err
	}

	return deletedID, nil
}
