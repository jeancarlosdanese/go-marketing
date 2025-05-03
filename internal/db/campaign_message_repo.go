// File: internal/db/campaign_message_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type CampaignMessageRepository interface {
	Create(ctx context.Context, msg *models.CampaignMessage) (*models.CampaignMessage, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.CampaignMessage, error)
	GetAllByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*models.CampaignMessage, error)
	Update(ctx context.Context, msg *models.CampaignMessage) error
	Delete(ctx context.Context, id uuid.UUID) error
	Approve(ctx context.Context, id uuid.UUID) error
}
