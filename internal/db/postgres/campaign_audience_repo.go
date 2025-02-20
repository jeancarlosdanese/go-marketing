// File: /internal/postgres/campaign_audience_repo.go

package postgres

import (
	"database/sql"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type campaignAudienceRepo struct {
	log *slog.Logger
	db  *sql.DB
}

func NewCampaignAudienceRepository(db *sql.DB) db.CampaignAudienceRepository {
	log := logger.GetLogger()

	return &campaignAudienceRepo{log: log, db: db}
}

// Adiciona contatos à audiência da campanha
func (r *campaignAudienceRepo) AddContactsToCampaign(campaignID uuid.UUID, contacts []models.CampaignAudience) ([]models.CampaignAudience, error) {
	audiences := []models.CampaignAudience{}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO campaigns_audience (campaign_id, contact_id, type, status)
		VALUES ($1, $2, $3, 'pendente') ON CONFLICT DO NOTHING RETURNING *
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	for _, contact := range contacts {
		var audience models.CampaignAudience
		err := stmt.QueryRow(campaignID, contact.ContactID, contact.Type).Scan(
			&audience.ID, &audience.CampaignID, &audience.ContactID, &audience.Type, &audience.Status, &audience.MessageID, &audience.Feedback, &audience.CreatedAt, &audience.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		audiences = append(audiences, audience)
	}

	tx.Commit()

	return audiences, nil
}

// Retorna a audiência de uma campanha, podendo filtrar por tipo (email ou whatsapp)
func (r *campaignAudienceRepo) GetCampaignAudience(campaignID uuid.UUID, contactType *string) ([]models.CampaignAudience, error) {
	query := `SELECT id, campaign_id, contact_id, "type", status, message_id, feedback_api, created_at, updated_at
FROM campaigns_audience WHERE campaign_id = $1`
	args := []interface{}{campaignID}

	if contactType != nil {
		query += " AND type = $2"
		args = append(args, *contactType)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var contacts []models.CampaignAudience
	for rows.Next() {
		var contact models.CampaignAudience
		if err := rows.Scan(&contact.ID, &contact.CampaignID, &contact.ContactID, &contact.Type, &contact.Status, &contact.MessageID, &contact.Feedback, &contact.CreatedAt, &contact.UpdatedAt); err != nil {
			return nil, err
		}
		contacts = append(contacts, contact)
	}
	return contacts, nil
}

// Remove um contato da audiência
func (r *campaignAudienceRepo) RemoveContactFromCampaign(campaignID, contactID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM campaigns_audience WHERE campaign_id = $1 AND contact_id = $2`, campaignID, contactID)
	return err
}

// Atualiza o status de um contato enviado
func (r *campaignAudienceRepo) UpdateStatus(contactID uuid.UUID, status, messageID string, feedback map[string]interface{}) error {
	feedbackJSON, err := json.Marshal(feedback)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		UPDATE campaigns_audience SET status = $1, message_id = $2, feedback_api = $3 WHERE contact_id = $4
	`, status, messageID, feedbackJSON, contactID)
	return err
}
