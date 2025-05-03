// File: /internal/db/campaign_audience_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type CampaignAudienceRepository interface {
	AddContactsToCampaign(ctx context.Context, campaignID uuid.UUID, contacts []models.CampaignAudience) ([]models.CampaignAudience, error)
	AddAllFilteredContacts(ctx context.Context, accountID uuid.UUID, campaignID uuid.UUID, filters *map[string]string, channelType models.ChannelType) error
	GetCampaignAudience(ctx context.Context, campaignID uuid.UUID, contactType *string) ([]dto.CampaignAudienceDTO, error)
	GetCampaignAudienceToSQS(ctx context.Context, accountID uuid.UUID, campaignID uuid.UUID, contactType *string) ([]dto.CampaignMessageDTO, error)
	RemoveContactFromCampaign(ctx context.Context, campaignID, audienceID uuid.UUID) error
	UpdateStatus(ctx context.Context, contactID uuid.UUID, status, messageID string, feedback map[string]interface{}) error
	UpdateStatusByMessageID(ctx context.Context, messageID string, status string, feedbackAPI *string) error
	GetPaginatedCampaignAudience(ctx context.Context, campaignID uuid.UUID, contactType *string, currentPage int, perPage int) (*models.Paginator, error)
	RemoveAllContactsFromCampaign(ctx context.Context, campaignID uuid.UUID) error
	GetRandomContact(ctx context.Context, campaignID uuid.UUID, channel string) (*models.Contact, error)
}
