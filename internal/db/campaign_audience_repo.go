// File: /internal/db/campaign_audience_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type CampaignAudienceRepository interface {
	AddContactsToCampaign(campaignID uuid.UUID, contacts []models.CampaignAudience) ([]models.CampaignAudience, error)
	GetCampaignAudience(campaignID uuid.UUID, contactType *string) ([]dto.CampaignMessageDTO, error)
	RemoveContactFromCampaign(campaignID, contactID uuid.UUID) error
	UpdateStatus(contactID uuid.UUID, status, messageID string, feedback map[string]interface{}) error
	UpdateStatusByMessageID(messageID string, status string, feedbackAPI *string) error
}
