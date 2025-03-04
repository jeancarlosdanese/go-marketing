// File: /internal/db/contact_import_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type ContactImportRepository interface {
	Create(ctx context.Context, importData *models.ContactImport) (*models.ContactImport, error)
	GetAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.ContactImport, error)
	GetByID(ctx context.Context, accountID uuid.UUID, id uuid.UUID) (*models.ContactImport, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateConfig(ctx context.Context, accountID uuid.UUID, id uuid.UUID, config models.ContactImportConfig) (*models.ContactImport, error)
}
