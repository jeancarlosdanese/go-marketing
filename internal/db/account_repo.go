// File: internal/db/account_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// AccountRepository define a interface para qualquer banco de dados
type AccountRepository interface {
	Create(ctx context.Context, account *models.Account) (*models.Account, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error)
	GetAll(ctx context.Context) ([]*models.Account, error)
	UpdateByID(ctx context.Context, id uuid.UUID, jsonData []byte) (*models.Account, error)
	DeleteByID(ctx context.Context, id uuid.UUID) error
}
