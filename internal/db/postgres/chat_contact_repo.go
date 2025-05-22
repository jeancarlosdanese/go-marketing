// internal/db/postgres/chat_contact_repo.go

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type chatContactRepository struct {
	db *sql.DB
}

func NewChatContactRepository(db *sql.DB) db.ChatContactRepository {
	return &chatContactRepository{db: db}
}

// ✅ Busca atendimento existente ou cria um novo para o contato no chat
func (r *chatContactRepository) FindOrCreate(ctx context.Context, accountID, chatID, contactID uuid.UUID) (*models.ChatContact, error) {
	// Primeiro tenta encontrar
	query := `
		SELECT id, account_id, chat_id, contact_id, status, created_at, updated_at
		FROM chat_contacts
		WHERE account_id = $1 AND chat_id = $2 AND contact_id = $3
		LIMIT 1
	`

	var found models.ChatContact
	err := r.db.QueryRowContext(ctx, query, accountID, chatID, contactID).Scan(
		&found.ID,
		&found.AccountID,
		&found.ChatID,
		&found.ContactID,
		&found.Status,
		&found.CreatedAt,
		&found.UpdatedAt,
	)
	if err == nil {
		return &found, nil
	}

	// Se não encontrou, cria novo
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	now := time.Now()
	newContact := &models.ChatContact{
		ID:        uuid.NewString(),
		AccountID: accountID.String(),
		ChatID:    chatID.String(),
		ContactID: contactID.String(),
		Status:    "aberto",
		CreatedAt: now,
		UpdatedAt: now,
	}

	insertQuery := `
		INSERT INTO chat_contacts (
			id, account_id, chat_id, contact_id, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = r.db.ExecContext(ctx, insertQuery,
		newContact.ID,
		newContact.AccountID,
		newContact.ChatID,
		newContact.ContactID,
		newContact.Status,
		newContact.CreatedAt,
		newContact.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return newContact, nil
}

// ListByChatID retorna todos os contatos de um chat
func (r *chatContactRepository) ListByChatID(ctx context.Context, accountID, chatID uuid.UUID) ([]*models.ChatContact, error) {
	query := `
		SELECT id, account_id, chat_id, contact_id, status, created_at, updated_at
		FROM chat_contacts
		WHERE account_id = $1 AND chat_id = $2
		ORDER BY updated_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contatos []*models.ChatContact
	for rows.Next() {
		var cc models.ChatContact
		if err := rows.Scan(
			&cc.ID,
			&cc.AccountID,
			&cc.ChatID,
			&cc.ContactID,
			&cc.Status,
			&cc.CreatedAt,
			&cc.UpdatedAt,
		); err != nil {
			return nil, err
		}
		contatos = append(contatos, &cc)
	}

	return contatos, nil
}
