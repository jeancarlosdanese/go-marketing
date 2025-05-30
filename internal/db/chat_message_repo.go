// internal/db/chat_message_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type ChatMessageRepository interface {
	Create(ctx context.Context, msg models.ChatMessage) (*models.ChatMessage, error)
	ListByChatContact(ctx context.Context, chatContactID uuid.UUID) ([]models.ChatMessage, error)
}
