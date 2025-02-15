// File: /internal/db/account_settings_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// AccountSettingsRepository define as operações para manipular as configurações de conta.
type AccountSettingsRepository interface {
	Create(settings *models.AccountSettings) (*models.AccountSettings, error)
	GetByAccountID(accountID uuid.UUID) (*models.AccountSettings, error)
	UpdateByAccountID(accountID uuid.UUID, settings *models.AccountSettings) (*models.AccountSettings, error)
	DeleteByAccountID(accountID uuid.UUID) error
}
