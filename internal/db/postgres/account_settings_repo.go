// File: /internal/db/postgres/account_settings_repo.go

package postgres

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// accountSettingsRepo implementa AccountSettingsRepository para PostgreSQL.
type accountSettingsRepo struct {
	db *sql.DB
}

// NewAccountSettingsRepository cria um novo repositório para configurações de conta.
func NewAccountSettingsRepository(db *sql.DB) *accountSettingsRepo {
	return &accountSettingsRepo{db: db}
}

// Create insere novas configurações para uma conta.
func (r *accountSettingsRepo) Create(settings *models.AccountSettings) (*models.AccountSettings, error) {
	query := `
		INSERT INTO account_settings (
			account_id, openai_api_key, evolution_instance, aws_access_key_id,
			aws_secret_access_key, aws_region, mail_from, mail_admin_to
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	err := r.db.QueryRow(
		query, settings.AccountID, settings.OpenAIAPIKey, settings.EvolutionInstance,
		settings.AWSAccessKeyID, settings.AWSSecretAccessKey, settings.AWSRegion,
		settings.MailFrom, settings.MailAdminTo,
	).Scan(&settings.ID)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar configurações da conta: %w", err)
	}

	return settings, nil
}

// GetByAccountID retorna as configurações de uma conta pelo ID da conta.
func (r *accountSettingsRepo) GetByAccountID(accountID uuid.UUID) (*models.AccountSettings, error) {
	query := `
		SELECT id, account_id, openai_api_key, evolution_instance, aws_access_key_id,
		       aws_secret_access_key, aws_region, mail_from, mail_admin_to
		FROM account_settings WHERE account_id = $1
	`
	settings := &models.AccountSettings{}

	err := r.db.QueryRow(query, accountID).Scan(
		&settings.ID, &settings.AccountID, &settings.OpenAIAPIKey, &settings.EvolutionInstance,
		&settings.AWSAccessKeyID, &settings.AWSSecretAccessKey, &settings.AWSRegion,
		&settings.MailFrom, &settings.MailAdminTo,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Retorna nil se não houver configurações para essa conta
		}
		return nil, fmt.Errorf("erro ao buscar configurações da conta: %w", err)
	}

	return settings, nil
}

// UpdateByAccountID atualiza as configurações de uma conta pelo ID da conta.
func (r *accountSettingsRepo) UpdateByAccountID(accountID uuid.UUID, settings *models.AccountSettings) (*models.AccountSettings, error) {
	query := `
		UPDATE account_settings
		SET openai_api_key = $1, evolution_instance = $2, aws_access_key_id = $3,
		    aws_secret_access_key = $4, aws_region = $5, mail_from = $6, mail_admin_to = $7
		WHERE account_id = $8
		RETURNING id
	`

	err := r.db.QueryRow(
		query, settings.OpenAIAPIKey, settings.EvolutionInstance, settings.AWSAccessKeyID,
		settings.AWSSecretAccessKey, settings.AWSRegion, settings.MailFrom, settings.MailAdminTo,
		accountID,
	).Scan(&settings.ID)

	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar configurações da conta: %w", err)
	}

	return settings, nil
}

// DeleteByAccountID remove as configurações de uma conta pelo ID da conta.
func (r *accountSettingsRepo) DeleteByAccountID(accountID uuid.UUID) error {
	query := `DELETE FROM account_settings WHERE account_id = $1`
	_, err := r.db.Exec(query, accountID)
	if err != nil {
		return fmt.Errorf("erro ao deletar configurações da conta: %w", err)
	}
	return nil
}
