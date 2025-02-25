// File: /internal/db/campaign_settings_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type CampaignSettingsRepository interface {
	CreateSettings(ctx context.Context, settings models.CampaignSettings) (*models.CampaignSettings, error)
	GetSettingsByCampaignID(ctx context.Context, campaignID uuid.UUID) (*models.CampaignSettings, error)
	UpdateSettings(ctx context.Context, settings models.CampaignSettings) (*models.CampaignSettings, error)
	DeleteSettings(ctx context.Context, campaignID uuid.UUID) error
	GetLastSettings(ctx context.Context, accountID uuid.UUID) (*models.CampaignSettings, error)
}
