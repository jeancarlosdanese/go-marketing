// internal/db/postgres/chat_message_repo.go

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

type chatMessageRepository struct {
	log *slog.Logger
	db  *sql.DB
}

func NewChatMessageRepository(db *sql.DB) db.ChatMessageRepository {
	return &chatMessageRepository{log: logger.GetLogger(), db: db}
}

// ✅ Create uma nova mensagem no histórico
func (r *chatMessageRepository) Create(ctx context.Context, msg models.ChatMessage) (*models.ChatMessage, error) {
	r.log.Debug("Criando nova mensagem no histórico", slog.Any("mensagem", msg))

	query := `
		INSERT INTO chat_messages (chat_contact_id, actor, type, content, file_url, source_processed)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, chat_contact_id, actor, type, content, file_url, source_processed, created_at, updated_at, deleted_at
	`
	var newMsg models.ChatMessage
	err := r.db.QueryRowContext(ctx, query,
		msg.ChatContactID,
		msg.Actor,
		msg.Type,
		msg.Content,
		msg.FileURL,
		msg.SourceProcessed,
	).Scan(
		&newMsg.ID,
		&newMsg.ChatContactID,
		&newMsg.Actor,
		&newMsg.Type,
		&newMsg.Content,
		&newMsg.FileURL,
		&newMsg.SourceProcessed,
		&newMsg.CreatedAt,
		&newMsg.UpdatedAt,
		&newMsg.DeletedAt,
	)
	if err != nil {
		r.log.Error("Erro ao inserir nova mensagem no histórico", slog.Any("erro", err))
		return nil, err
	}

	r.log.Debug("Mensagem criada com sucesso", slog.Any("mensagem", newMsg))
	return &newMsg, nil
}

// ListByChatContact retorna todas as mensagens de um contato específico
func (r *chatMessageRepository) ListByChatContact(ctx context.Context, chatContactID uuid.UUID) ([]models.ChatMessage, error) {
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

	var messages []models.ChatMessage
	for rows.Next() {
		var message models.ChatMessage
		err := rows.Scan(
			&message.ID,
			&message.ChatContactID,
			&message.Actor,
			&message.Type,
			&message.Content,
			&message.FileURL,
			&message.SourceProcessed,
			&message.CreatedAt,
			&message.UpdatedAt,
			&message.DeletedAt,
		)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}
