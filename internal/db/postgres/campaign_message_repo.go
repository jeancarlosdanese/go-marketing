// File: internal/db/postgres/campaign_message_repo.go

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type campaignMessageRepository struct {
	db *sql.DB
}

func NewCampaignMessageRepository(db *sql.DB) db.CampaignMessageRepository {
	return &campaignMessageRepository{db: db}
}

func (r *campaignMessageRepository) Create(ctx context.Context, msg *models.CampaignMessage) (*models.CampaignMessage, error) {
	query := `
		INSERT INTO campaign_messages (
			campaign_id, contact_id, channel,
			saudacao, corpo, finalizacao, assinatura,
			prompt_usado, feedback, version, is_approved, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now(), now())
		RETURNING id, created_at, updated_at
	`

	feedbackJSON, _ := json.Marshal(msg.Feedback)

	err := r.db.QueryRowContext(ctx, query,
		msg.CampaignID,
		msg.ContactID,
		msg.Channel,
		msg.Saudacao,
		msg.Corpo,
		msg.Finalizacao,
		msg.Assinatura,
		msg.PromptUsado,
		feedbackJSON,
		msg.Version,
		msg.IsApproved,
	).Scan(&msg.ID, &msg.CreatedAt, &msg.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (r *campaignMessageRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.CampaignMessage, error) {
	query := `SELECT id, campaign_id, contact_id, channel, saudacao, corpo, finalizacao, assinatura, prompt_usado, feedback, version, is_approved, created_at, updated_at FROM campaign_messages WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	var msg models.CampaignMessage
	var feedbackJSON []byte

	err := row.Scan(
		&msg.ID, &msg.CampaignID, &msg.ContactID, &msg.Channel,
		&msg.Saudacao, &msg.Corpo, &msg.Finalizacao, &msg.Assinatura,
		&msg.PromptUsado, &feedbackJSON, &msg.Version, &msg.IsApproved,
		&msg.CreatedAt, &msg.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(feedbackJSON, &msg.Feedback)
	return &msg, nil
}

func (r *campaignMessageRepository) GetAllByCampaignID(ctx context.Context, campaignID uuid.UUID) ([]*models.CampaignMessage, error) {
	query := `SELECT id, campaign_id, contact_id, channel, saudacao, corpo, finalizacao, assinatura, prompt_usado, feedback, version, is_approved, created_at, updated_at FROM campaign_messages WHERE campaign_id = $1`

	rows, err := r.db.QueryContext(ctx, query, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*models.CampaignMessage

	for rows.Next() {
		var msg models.CampaignMessage
		var feedbackJSON []byte

		err := rows.Scan(
			&msg.ID, &msg.CampaignID, &msg.ContactID, &msg.Channel,
			&msg.Saudacao, &msg.Corpo, &msg.Finalizacao, &msg.Assinatura,
			&msg.PromptUsado, &feedbackJSON, &msg.Version, &msg.IsApproved,
			&msg.CreatedAt, &msg.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		_ = json.Unmarshal(feedbackJSON, &msg.Feedback)
		messages = append(messages, &msg)
	}

	return messages, nil
}

func (r *campaignMessageRepository) Update(ctx context.Context, msg *models.CampaignMessage) error {
	query := `
		UPDATE campaign_messages SET
			saudacao = $1, corpo = $2, finalizacao = $3, assinatura = $4,
			prompt_usado = $5, feedback = $6, version = $7, is_approved = $8, updated_at = now()
		WHERE id = $9
	`

	feedbackJSON, _ := json.Marshal(msg.Feedback)

	_, err := r.db.ExecContext(ctx, query,
		msg.Saudacao,
		msg.Corpo,
		msg.Finalizacao,
		msg.Assinatura,
		msg.PromptUsado,
		feedbackJSON,
		msg.Version,
		msg.IsApproved,
		msg.ID,
	)

	return err
}

func (r *campaignMessageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM campaign_messages WHERE id = $1`, id)
	return err
}

func (r *campaignMessageRepository) Approve(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE campaign_messages SET is_approved = true, updated_at = now() WHERE id = $1`, id)
	return err
}

func (r *campaignMessageRepository) GetNextVersion(ctx context.Context, campaignID uuid.UUID, channel string, contactID *uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(version), 0) + 1 
		FROM campaign_messages 
		WHERE campaign_id = $1 AND channel = $2 AND contact_id IS NOT DISTINCT FROM $3;
	`

	var nextVersion int
	err := r.db.QueryRowContext(ctx, query, campaignID, channel, contactID).Scan(&nextVersion)
	if err != nil {
		return 0, err
	}

	return nextVersion, nil
}
