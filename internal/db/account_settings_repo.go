// File: /internal/db/account_settings_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// AccountSettingsRepository define as operações para manipular as configurações de conta.
type AccountSettingsRepository interface {
	Create(ctx context.Context, settings *models.AccountSettings) (*models.AccountSettings, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) (*models.AccountSettings, error)
	UpdateByAccountID(ctx context.Context, accountID uuid.UUID, settings *models.AccountSettings) (*models.AccountSettings, error)
	DeleteByAccountID(ctx context.Context, accountID uuid.UUID) error
}
