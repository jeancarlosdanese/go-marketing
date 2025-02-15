// File: /internal/db/template_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// TemplateRepository define as operações para manipular templates.
type TemplateRepository interface {
	Create(template *models.Template) (*models.Template, error)
	GetByID(templateID uuid.UUID) (*models.Template, error)
	GetByAccountID(accountID uuid.UUID) ([]models.Template, error)
	UpdateByID(templateID uuid.UUID, template *models.Template) (*models.Template, error)
	DeleteByID(templateID uuid.UUID) error
}
