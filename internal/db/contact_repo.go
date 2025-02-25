// File: /internal/db/contact_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// ContactRepository define as operações para manipular contatos.
type ContactRepository interface {
	Create(ctx context.Context, contact *models.Contact) (*models.Contact, error)
	GetByID(ctx context.Context, contactID uuid.UUID) (*models.Contact, error)
	GetByAccountID(ctx context.Context, accountID uuid.UUID, filters map[string]string) ([]models.Contact, error)
	FindByEmailOrWhatsApp(ctx context.Context, accountID uuid.UUID, email, whatsapp *string) (*models.Contact, error)
	UpdateByID(ctx context.Context, contactID uuid.UUID, contact *models.Contact) (*models.Contact, error)
	DeleteByID(ctx context.Context, contactID uuid.UUID) error
}
