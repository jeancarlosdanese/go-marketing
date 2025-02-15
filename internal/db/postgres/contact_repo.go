// File: /internal/db/postgres/contact_repo.go

package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// contactRepo implementa ContactRepository para PostgreSQL.
type contactRepo struct {
	log *slog.Logger
	db  *sql.DB
}

// NewContactRepository cria um novo repositório para contatos.
func NewContactRepository(db *sql.DB) *contactRepo {
	log := logger.GetLogger()
	return &contactRepo{log: log, db: db}
}

// Create insere um novo contato no banco de dados.
func (r *contactRepo) Create(contact *models.Contact) (*models.Contact, error) {
	r.log.Debug("Criando novo contato", "name", contact.Name, "email", contact.Email)
	query := `
		INSERT INTO contacts (
			id, account_id, name, email, whatsapp, gender, birth_date, history, opt_out_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	contact.ID = uuid.New()
	err := r.db.QueryRow(
		query,
		contact.ID, contact.AccountID, contact.Name, contact.Email, contact.WhatsApp,
		contact.Gender, contact.BirthDate, contact.History, contact.OptOutAt,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar contato: %w", err)
	}

	r.log.Debug("Contato criado com sucesso", "id", contact.ID)

	return contact, nil
}

// GetByID retorna um contato pelo ID.
func (r *contactRepo) GetByID(contactID uuid.UUID) (*models.Contact, error) {
	r.log.Debug("Buscando contato por ID", "id", contactID)

	query := `
		SELECT id, account_id, name, email, whatsapp, gender, birth_date, history, opt_out_at, created_at, updated_at
		FROM contacts WHERE id = $1
	`
	contact := &models.Contact{}
	err := r.db.QueryRow(query, contactID).Scan(
		&contact.ID, &contact.AccountID, &contact.Name, &contact.Email, &contact.WhatsApp,
		&contact.Gender, &contact.BirthDate, &contact.History, &contact.OptOutAt,
		&contact.CreatedAt, &contact.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Retorna nil se não encontrar
		}
		return nil, fmt.Errorf("erro ao buscar contato: %w", err)
	}

	r.log.Debug("Contato encontrado", "id", contact.ID)

	return contact, nil
}

// GetByAccountID retorna todos os contatos de uma conta específica.
func (r *contactRepo) GetByAccountID(accountID uuid.UUID) ([]models.Contact, error) {
	r.log.Debug("Buscando contatos por account_id", "account_id", accountID)

	query := `
		SELECT id, account_id, name, email, whatsapp, gender, birth_date, history, opt_out_at, created_at, updated_at
		FROM contacts WHERE account_id = $1
	`
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar contatos: %w", err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		if err := rows.Scan(
			&contact.ID, &contact.AccountID, &contact.Name, &contact.Email, &contact.WhatsApp,
			&contact.Gender, &contact.BirthDate, &contact.History, &contact.OptOutAt,
			&contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear contatos: %w", err)
		}
		contacts = append(contacts, contact)
	}

	r.log.Debug("Contatos encontrados", "total", len(contacts))

	return contacts, nil
}

// UpdateByID atualiza um contato pelo ID.
func (r *contactRepo) UpdateByID(contactID uuid.UUID, contact *models.Contact) (*models.Contact, error) {
	r.log.Debug("Atualizando contato por ID", "id", contactID)

	query := `
		UPDATE contacts
		SET name = $1, email = $2, whatsapp = $3, gender = $4, birth_date = $5, history = $6, opt_out_at = $7, updated_at = NOW()
		WHERE id = $8
		RETURNING updated_at
	`
	err := r.db.QueryRow(
		query,
		contact.Name, contact.Email, contact.WhatsApp, contact.Gender, contact.BirthDate,
		contact.History, contact.OptOutAt, contactID,
	).Scan(&contact.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar contato: %w", err)
	}

	r.log.Debug("Contato atualizado com sucesso", "id", contactID)

	return contact, nil
}

// DeleteByID remove um contato pelo ID.
func (r *contactRepo) DeleteByID(contactID uuid.UUID) error {
	r.log.Debug("Deletando contato por ID", "id", contactID)

	query := `DELETE FROM contacts WHERE id = $1`
	_, err := r.db.Exec(query, contactID)
	if err != nil {
		return fmt.Errorf("erro ao deletar contato: %w", err)
	}

	r.log.Debug("Contato deletado com sucesso", "id", contactID)

	return nil
}
