// File: /internal/postgres/campaign_settings_repo.go

package postgres

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// campaignSettingsRepository gerencia as configurações das campanhas no banco
type campaignSettingsRepository struct {
	log *slog.Logger
	db  *sql.DB
}

// NewCampaignSettingsRepository cria um novo repositório
func NewCampaignSettingsRepository(db *sql.DB) db.CampaignSettingsRepository {
	log := logger.GetLogger()
	return &campaignSettingsRepository{log: log, db: db}
}

// ✅ Criar configurações para uma campanha
func (r *campaignSettingsRepository) CreateSettings(ctx context.Context, settings models.CampaignSettings) (*models.CampaignSettings, error) {
	r.log.Debug("📩 Criando configurações da campanha", "campaign_id", settings.CampaignID)

	query := `
		INSERT INTO campaign_settings (
			campaign_id, brand, subject, tone, email_from, email_reply, 
			email_footer, email_instructions, whatsapp_from, whatsapp_reply, 
			whatsapp_footer, whatsapp_instructions
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		settings.CampaignID, settings.Brand, settings.Subject, settings.Tone,
		settings.EmailFrom, settings.EmailReply, settings.EmailFooter, settings.EmailInstructions,
		settings.WhatsAppFrom, settings.WhatsAppReply, settings.WhatsAppFooter, settings.WhatsAppInstructions,
	).Scan(&settings.ID, &settings.CreatedAt, &settings.UpdatedAt)

	if err != nil {
		r.log.Error("❌ Erro ao criar configurações da campanha", "error", err)
		return nil, err
	}

	r.log.Info("✅ Configurações da campanha criadas com sucesso", "campaign_id", settings.CampaignID)
	return &settings, nil
}

// ✅ Buscar configurações de uma campanha pelo `campaign_id`
func (r *campaignSettingsRepository) GetSettingsByCampaignID(ctx context.Context, campaignID uuid.UUID) (*models.CampaignSettings, error) {
	r.log.Debug("📩 Buscando configurações da campanha", "campaign_id", campaignID)

	query := `
		SELECT id, campaign_id, brand, subject, tone, email_from, email_reply, 
			   email_footer, email_instructions, whatsapp_from, whatsapp_reply, 
			   whatsapp_footer, whatsapp_instructions, created_at, updated_at
		FROM campaign_settings
		WHERE campaign_id = $1
	`
	settings := models.CampaignSettings{}

	err := r.db.QueryRowContext(ctx, query, campaignID).Scan(
		&settings.ID, &settings.CampaignID, &settings.Brand, &settings.Subject, &settings.Tone,
		&settings.EmailFrom, &settings.EmailReply, &settings.EmailFooter, &settings.EmailInstructions,
		&settings.WhatsAppFrom, &settings.WhatsAppReply, &settings.WhatsAppFooter, &settings.WhatsAppInstructions,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warn("⚠️ Nenhuma configuração encontrada para a campanha", "campaign_id", campaignID)
			return nil, nil
		}
		r.log.Error("❌ Erro ao buscar configurações da campanha", "error", err)
		return nil, err
	}

	r.log.Info("✅ Configurações da campanha encontradas", "campaign_id", campaignID)
	return &settings, nil
}

// ✅ Atualizar configurações da campanha
func (r *campaignSettingsRepository) UpdateSettings(ctx context.Context, settings models.CampaignSettings) (*models.CampaignSettings, error) {
	r.log.Debug("📩 Atualizando configurações da campanha", "campaign_id", settings.CampaignID)

	query := `
		UPDATE campaign_settings
		SET brand = $2, subject = $3, tone = $4, email_from = $5, email_reply = $6,
			email_footer = $7, email_instructions = $8, whatsapp_from = $9, 
			whatsapp_reply = $10, whatsapp_footer = $11, whatsapp_instructions = $12,
			updated_at = now()
		WHERE campaign_id = $1
		RETURNING id, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		settings.CampaignID, settings.Brand, settings.Subject, settings.Tone,
		settings.EmailFrom, settings.EmailReply, settings.EmailFooter, settings.EmailInstructions,
		settings.WhatsAppFrom, settings.WhatsAppReply, settings.WhatsAppFooter, settings.WhatsAppInstructions,
	).Scan(&settings.ID, &settings.UpdatedAt)

	if err != nil {
		r.log.Error("❌ Erro ao atualizar configurações da campanha", "error", err)
		return nil, err
	}

	r.log.Info("✅ Configurações da campanha atualizadas com sucesso", "campaign_id", settings.CampaignID)
	return &settings, nil
}

// ✅ Excluir configurações da campanha pelo `campaign_id`
func (r *campaignSettingsRepository) DeleteSettings(ctx context.Context, campaignID uuid.UUID) error {
	r.log.Debug("🗑️ Excluindo configurações da campanha", "campaign_id", campaignID)

	query := `DELETE FROM campaign_settings WHERE campaign_id = $1`
	result, err := r.db.ExecContext(ctx, query, campaignID)

	if err != nil {
		r.log.Error("❌ Erro ao excluir configurações da campanha", "error", err)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		r.log.Warn("⚠️ Nenhuma configuração excluída, campanha não encontrada", "campaign_id", campaignID)
		return nil
	}

	r.log.Info("🗑️ Configurações da campanha excluídas", "campaign_id", campaignID)
	return nil
}

// ✅ Buscar a última configuração utilizada por uma conta (`account_id`)
func (r *campaignSettingsRepository) GetLastSettings(ctx context.Context, accountID uuid.UUID) (*models.CampaignSettings, error) {
	r.log.Debug("📩 Buscando última configuração utilizada pela conta", "account_id", accountID)

	query := `
		SELECT cs.id, cs.campaign_id, cs.brand, cs.subject, cs.tone, 
			   cs.email_from, cs.email_reply, cs.email_footer, cs.email_instructions, 
			   cs.whatsapp_from, cs.whatsapp_reply, cs.whatsapp_footer, cs.whatsapp_instructions, 
			   cs.created_at, cs.updated_at
		FROM campaign_settings cs
		JOIN campaigns c ON cs.campaign_id = c.id
		WHERE c.account_id = $1
		ORDER BY cs.created_at DESC
		LIMIT 1
	`

	settings := models.CampaignSettings{}

	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&settings.ID, &settings.CampaignID, &settings.Brand, &settings.Subject, &settings.Tone,
		&settings.EmailFrom, &settings.EmailReply, &settings.EmailFooter, &settings.EmailInstructions,
		&settings.WhatsAppFrom, &settings.WhatsAppReply, &settings.WhatsAppFooter, &settings.WhatsAppInstructions,
		&settings.CreatedAt, &settings.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.log.Warn("⚠️ Nenhuma configuração recente encontrada para a conta", "account_id", accountID)
			return nil, nil
		}
		r.log.Error("❌ Erro ao buscar última configuração da conta", "error", err)
		return nil, err
	}

	r.log.Info("✅ Última configuração da conta encontrada", "account_id", accountID)
	return &settings, nil
}
