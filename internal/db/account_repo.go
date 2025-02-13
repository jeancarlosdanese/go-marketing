// File: internal/db/account_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// AccountRepository define a interface para qualquer banco de dados
type AccountRepository interface {
	Create(account *models.Account) (*models.Account, error)
	GetByID(id uuid.UUID) (*models.Account, error)
}
