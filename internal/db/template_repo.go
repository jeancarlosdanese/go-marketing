// File: /internal/db/template_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// TemplateRepository define as operações para manipular templates.
type TemplateRepository interface {
	Create(ctx context.Context, template *models.Template) (*models.Template, error)
	GetByID(ctx context.Context, templateID uuid.UUID) (*models.Template, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.Template, error)
	UpdateByID(ctx context.Context, templateID uuid.UUID, template *models.Template) (*models.Template, error)
	DeleteByID(ctx context.Context, templateID uuid.UUID) error
}
