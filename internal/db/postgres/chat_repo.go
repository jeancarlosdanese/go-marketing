// internal/db/postgres/chat_repo.go

package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type chatRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewChatRepository(db *sql.DB) db.ChatRepository {
	return &chatRepository{log: logger.GetLogger(), db: db}
}

// Insert creates a new chat in the database.
func (r *chatRepository) Insert(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	query := `
		INSERT INTO chats (
			id, account_id, department, title, instructions,
			phone_number, evolution_instance, webhook_url,
			status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11
		)
		RETURNING id, account_id, department, title, instructions,
		          phone_number, evolution_instance, webhook_url,
		          status, created_at, updated_at
	`

	var inserted models.Chat
	err := r.db.QueryRowContext(ctx, query,
		chat.ID,
		chat.AccountID,
		chat.Department,
		chat.Title,
		chat.Instructions,
		chat.PhoneNumber,
		chat.EvolutionInstance,
		chat.WebhookURL,
		chat.Status,
		chat.CreatedAt,
		chat.UpdatedAt,
	).Scan(
		&inserted.ID,
		&inserted.AccountID,
		&inserted.Department,
		&inserted.Title,
		&inserted.Instructions,
		&inserted.PhoneNumber,
		&inserted.EvolutionInstance,
		&inserted.WebhookURL,
		&inserted.Status,
		&inserted.CreatedAt,
		&inserted.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}
	return &inserted, nil
}

// ListByAccountID retrieves all chats for a given account ID.
func (r *chatRepository) ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Chat, error) {
	query := `
		SELECT id, account_id, department, title, instructions,
		       phone_number, evolution_instance, webhook_url,
		       status, created_at, updated_at
		FROM chats
		WHERE account_id = $1
		ORDER BY title ASC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*models.Chat
	for rows.Next() {
		var chat models.Chat
		if err := rows.Scan(
			&chat.ID,
			&chat.AccountID,
			&chat.Department,
			&chat.Title,
			&chat.Instructions,
			&chat.PhoneNumber,
			&chat.EvolutionInstance,
			&chat.WebhookURL,
			&chat.Status,
			&chat.CreatedAt,
			&chat.UpdatedAt,
		); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}

	return chats, nil
}

// GetByID retrieves a chat by its ID and account ID.
func (r *chatRepository) GetByID(ctx context.Context, accountID, chatID uuid.UUID) (*models.Chat, error) {
	query := `
		SELECT id, account_id, department, title, instructions,
		       phone_number, evolution_instance, webhook_url,
		       status, created_at, updated_at
		FROM chats
		WHERE account_id = $1 AND id = $2
		LIMIT 1
	`

	var chat models.Chat
	err := r.db.QueryRowContext(ctx, query, accountID, chatID).Scan(
		&chat.ID,
		&chat.AccountID,
		&chat.Department,
		&chat.Title,
		&chat.Instructions,
		&chat.PhoneNumber,
		&chat.EvolutionInstance,
		&chat.WebhookURL,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &chat, nil
}

// GetActiveByID retrieves an active chat by its ID and account ID.
func (r *chatRepository) GetActiveByID(ctx context.Context, accountID, chatID uuid.UUID) (*models.Chat, error) {
	query := `
		SELECT id, account_id, department, title, instructions,
		       phone_number, evolution_instance, webhook_url,
		       status, created_at, updated_at
		FROM chats
		WHERE account_id = $1 AND id = $2 AND status = 'ativo'
		LIMIT 1
	`

	logger.Debug("Executing query: %s, with params: %v", query, []interface{}{accountID, chatID})

	var chat models.Chat
	err := r.db.QueryRowContext(ctx, query, accountID, chatID).Scan(
		&chat.ID,
		&chat.AccountID,
		&chat.Department,
		&chat.Title,
		&chat.Instructions,
		&chat.PhoneNumber,
		&chat.EvolutionInstance,
		&chat.WebhookURL,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

// ✅ Retorna o chat ativo configurado para o setor (ex: financeiro)
func (r *chatRepository) GetActiveByDepartment(ctx context.Context, accountID, department string) (*models.Chat, error) {
	query := `
		SELECT id, account_id, department, title, instructions,
		       phone_number, evolution_instance, webhook_url,
		       status, created_at, updated_at
		FROM chats
		WHERE account_id = $1 AND department = $2 AND status = 'ativo'
		LIMIT 1
	`

	var chat models.Chat
	err := r.db.QueryRowContext(ctx, query, accountID, department).Scan(
		&chat.ID,
		&chat.AccountID,
		&chat.Department,
		&chat.Title,
		&chat.Instructions,
		&chat.PhoneNumber,
		&chat.EvolutionInstance,
		&chat.WebhookURL,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

// Update updates an existing chat in the database.
func (r *chatRepository) Update(ctx context.Context, chat *models.Chat) (*models.Chat, error) {
	query := `
		UPDATE chats
		SET title = $1,
		    instructions = $2,
		    phone_number = $3,
		    evolution_instance = $4,
		    webhook_url = $5,
		    updated_at = $6
		WHERE id = $7 AND account_id = $8
		RETURNING id, account_id, department, title, instructions,
		          phone_number, evolution_instance, webhook_url,
		          status, created_at, updated_at
	`

	var updated models.Chat
	err := r.db.QueryRowContext(ctx, query,
		chat.Title,
		chat.Instructions,
		chat.PhoneNumber,
		chat.EvolutionInstance,
		chat.WebhookURL,
		chat.UpdatedAt,
		chat.ID,
		chat.AccountID,
	).Scan(
		&updated.ID,
		&updated.AccountID,
		&updated.Department,
		&updated.Title,
		&updated.Instructions,
		&updated.PhoneNumber,
		&updated.EvolutionInstance,
		&updated.WebhookURL,
		&updated.Status,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

// GetActiveByEvolutionInstance retrieves an active chat by its Evolution instance.
func (r *chatRepository) GetActiveByEvolutionInstance(ctx context.Context, instance string) (*models.Chat, error) {
	query := `
		SELECT id, account_id, department, title, instructions, phone_number,
		       evolution_instance, webhook_url, status, created_at, updated_at
		FROM chats
		WHERE evolution_instance = $1 AND status = 'ativo'
		LIMIT 1
	`

	var chat models.Chat
	err := r.db.QueryRowContext(ctx, query, instance).Scan(
		&chat.ID,
		&chat.AccountID,
		&chat.Department,
		&chat.Title,
		&chat.Instructions,
		&chat.PhoneNumber,
		&chat.EvolutionInstance,
		&chat.WebhookURL,
		&chat.Status,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}
