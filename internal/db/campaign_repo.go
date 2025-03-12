// File: /internal/db/campaign_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// CampaignRepository define as operações para manipular campanhas
type CampaignRepository interface {
	Create(ctx context.Context, campaign *models.Campaign) (*models.Campaign, error)
	GetByID(ctx context.Context, campaignID uuid.UUID) (*models.Campaign, error)
	GetAllByAccountID(ctx context.Context, accountID uuid.UUID, filters *map[string]string) ([]models.Campaign, error)
	UpdateByID(ctx context.Context, campaignID uuid.UUID, campaign *models.Campaign) (*models.Campaign, error)
	UpdateStatus(ctx context.Context, campaignID uuid.UUID, status string) error
	DeleteByID(ctx context.Context, campaignID uuid.UUID) error
}
