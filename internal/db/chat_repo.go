// internal/db/chat_message_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type ChatRepository interface {
	Insert(ctx context.Context, chat *models.Chat) (*models.Chat, error)
	ListByAccountID(ctx context.Context, accountID uuid.UUID) ([]*models.Chat, error)
	GetByID(ctx context.Context, accountID, chatID uuid.UUID) (*models.Chat, error)
	Update(ctx context.Context, chat *models.Chat) (*models.Chat, error)
	GetActiveByID(ctx context.Context, accountID, chatID uuid.UUID) (*models.Chat, error)
	GetActiveByDepartment(ctx context.Context, accountID, department string) (*models.Chat, error)
	GetActiveByEvolutionInstance(ctx context.Context, instance string) (*models.Chat, error)
}
