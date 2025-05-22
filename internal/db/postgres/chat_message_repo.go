// internal/db/postgres/chat_message_repo.go

package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type chatMessageRepository struct {
	db *sql.DB
}

func NewChatMessageRepository(db *sql.DB) db.ChatMessageRepository {
	return &chatMessageRepository{db: db}
}

// ✅ Insere uma nova mensagem no histórico
func (r *chatMessageRepository) Insert(ctx context.Context, msg models.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (
			id, chat_contact_id, actor, type, content, file_url, source_processed, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		msg.ID,
		msg.ChatContactID,
		msg.Actor,
		msg.Type,
		msg.Content,
		msg.FileURL,
		msg.SourceProcessed,
		msg.CreatedAt,
		msg.UpdatedAt,
	)
	return err
}

// ListByChatContact retorna todas as mensagens de um contato específico
func (r *chatMessageRepository) ListByChatContact(ctx context.Context, chatContactID uuid.UUID) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, chat_contact_id, actor, type, content, file_url,
		       source_processed, created_at, updated_at, deleted_at
		FROM chat_messages
		WHERE chat_contact_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, chatContactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mensagens []*models.ChatMessage
	for rows.Next() {
		var m models.ChatMessage
		err := rows.Scan(
			&m.ID,
			&m.ChatContactID,
			&m.Actor,
			&m.Type,
			&m.Content,
			&m.FileURL,
			&m.SourceProcessed,
			&m.CreatedAt,
			&m.UpdatedAt,
			&m.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		mensagens = append(mensagens, &m)
	}

	return mensagens, nil
}

// ✅ Lista mensagens de um atendimento específico (ordenadas por data)
func (r *chatMessageRepository) GetByChatContact(ctx context.Context, chatContactID string) ([]models.ChatMessage, error) {
	query := `
		SELECT id, chat_contact_id, actor, type, content, file_url, source_processed, created_at, updated_at, deleted_at
		FROM chat_messages
		WHERE chat_contact_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, chatContactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []models.ChatMessage
	for rows.Next() {
		var msg models.ChatMessage
		err := rows.Scan(
			&msg.ID,
			&msg.ChatContactID,
			&msg.Actor,
			&msg.Type,
			&msg.Content,
			&msg.FileURL,
			&msg.SourceProcessed,
			&msg.CreatedAt,
			&msg.UpdatedAt,
			&msg.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
