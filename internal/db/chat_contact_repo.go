// internal/db/chat_message_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type ChatContactRepository interface {
	FindOrCreate(ctx context.Context, accountID, chatID, contactID uuid.UUID) (*models.ChatContact, error)
	ListByChatID(ctx context.Context, accountID, chatID uuid.UUID) ([]*models.ChatContact, error)
}
