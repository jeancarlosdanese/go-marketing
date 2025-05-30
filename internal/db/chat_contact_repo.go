// internal/db/chat_message_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type ChatContactRepository interface {
	FindOrCreate(ctx context.Context, accountID, chatID, whatsappContactID uuid.UUID) (*models.ChatContact, error)
	FindByID(ctx context.Context, accountID, chatID, chatContactID uuid.UUID) (*models.ChatContact, error)
	ListByChatID(ctx context.Context, accountID, chatID uuid.UUID) ([]dto.ChatContactFull, error)
}
