// internal/db/whatsapp_contact_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type WhatsappContactRepository interface {
	FindOrCreate(ctx context.Context, contact *models.WhatsappContact) (*models.WhatsappContact, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.WhatsappContact, error)
}
