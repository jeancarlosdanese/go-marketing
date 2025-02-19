// File: /internal/db/contact_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// ContactRepository define as operações para manipular contatos.
type ContactRepository interface {
	Create(contact *models.Contact) (*models.Contact, error)
	GetByID(contactID uuid.UUID) (*models.Contact, error)
	GetByAccountID(accountID uuid.UUID, filters map[string]string) ([]models.Contact, error)
	FindByEmailOrWhatsApp(accountID uuid.UUID, email, whatsapp *string) (*models.Contact, error)
	UpdateByID(contactID uuid.UUID, contact *models.Contact) (*models.Contact, error)
	DeleteByID(contactID uuid.UUID) error
}
