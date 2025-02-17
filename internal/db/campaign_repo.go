// File: /internal/db/campaign_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// CampaignRepository define as operações para manipular campanhas
type CampaignRepository interface {
	Create(campaign *models.Campaign) (*models.Campaign, error)
	GetByID(campaignID uuid.UUID) (*models.Campaign, error)
	GetAllByAccountID(accountID uuid.UUID) ([]models.Campaign, error)
	UpdateByID(campaignID uuid.UUID, campaign *models.Campaign) (*models.Campaign, error)
	UpdateStatus(campaignID uuid.UUID, status string) error
	DeleteByID(campaignID uuid.UUID) error
}
