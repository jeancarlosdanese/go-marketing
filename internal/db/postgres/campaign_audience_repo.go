// File: /internal/postgres/campaign_audience_repo.go

package postgres

import (
	"database/sql"
	"encoding/json"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
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

// GetCampaignAudience retorna a audiência de uma campanha junto com os detalhes dos contatos
func (r *campaignAudienceRepo) GetCampaignAudience(campaignID uuid.UUID, contactType *string) ([]dto.CampaignMessageDTO, error) {
	query := `
		SELECT ca.id, ca.campaign_id, ca.contact_id, ca.type, ca.status,
		       c.name, c.email, c.whatsapp, c.gender, c.birth_date, 
		       c.bairro, c.cidade, c.estado, c.tags, c.history, 
		       c.opt_out_at, c.last_contact_at, c.created_at, c.updated_at
		FROM campaigns_audience ca
		INNER JOIN contacts c ON ca.contact_id = c.id
		WHERE ca.campaign_id = $1`
	args := []interface{}{campaignID}

	if contactType != nil {
		query += " AND ca.type = $2"
		args = append(args, *contactType)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []dto.CampaignMessageDTO
	for rows.Next() {
		var msg dto.CampaignMessageDTO
		var tagsJSON []byte

		if err := rows.Scan(
			&msg.ID, &msg.CampaignID, &msg.ContactID, &msg.Type, &msg.Status,
			&msg.Name, &msg.Email, &msg.WhatsApp, &msg.Gender, &msg.BirthDate,
			&msg.Bairro, &msg.Cidade, &msg.Estado, &tagsJSON, &msg.History,
			&msg.OptOutAt, &msg.LastContactAt, &msg.CreatedAt, &msg.UpdatedAt,
		); err != nil {
			return nil, err
		}

		// Converter JSONB das tags para map[string]interface{}
		if err := json.Unmarshal(tagsJSON, &msg.Tags); err != nil {
			msg.Tags = make(map[string]interface{}) // Se der erro, inicializa vazio
		}

		messages = append(messages, msg)
	}

	return messages, nil
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

// UpdateStatusByMessageID atualiza o status de uma mensagem usando o message_id
func (r *campaignAudienceRepo) UpdateStatusByMessageID(messageID string, status string, feedbackAPI *string) error {
	query := `
		UPDATE campaigns_audience 
		SET status = $1, feedback_api = $2, updated_at = NOW()
		WHERE message_id = $3;
	`
	_, err := r.db.Exec(query, status, feedbackAPI, messageID)
	if err != nil {
		r.log.Error("❌ Erro ao atualizar status por message_id: %s, erro: %v", messageID, err)
		return err
	}

	r.log.Info("✅ Status atualizado com sucesso para message_id: %s -> %s", messageID, status)
	return nil
}
