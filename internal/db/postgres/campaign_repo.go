// File: /internal/db/postgres/campaign_repo.go

package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// campaignRepo implementa CampaignRepository para PostgreSQL
type campaignRepo struct {
	log *slog.Logger
	db  *sql.DB
}

// NewCampaignRepository cria um novo repositório para campanhas
func NewCampaignRepository(db *sql.DB) *campaignRepo {
	log := logger.GetLogger()
	return &campaignRepo{log: log, db: db}
}

// Create insere uma nova campanha no banco de dados
func (r *campaignRepo) Create(campaign *models.Campaign) (*models.Campaign, error) {
	r.log.Debug("Criando nova campanha", "name", campaign.Name)

	// Serializa Channels e Filters para JSONB
	channelsJSON, err := json.Marshal(campaign.Channels)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar channels: %w", err)
	}
	filtersJSON, err := json.Marshal(campaign.Filters)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar filters: %w", err)
	}

	query := `
		INSERT INTO campaigns (
			id, account_id, name, description, channels, filters, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	campaign.ID = uuid.New()

	err = r.db.QueryRow(
		query, campaign.ID, campaign.AccountID, campaign.Name, campaign.Description,
		channelsJSON, filtersJSON, campaign.Status,
	).Scan(&campaign.ID, &campaign.CreatedAt, &campaign.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar campanha: %w", err)
	}

	r.log.Debug("Campanha criada com sucesso", "id", campaign.ID)
	return campaign, nil
}

// GetByID retorna uma campanha pelo ID
func (r *campaignRepo) GetByID(campaignID uuid.UUID) (*models.Campaign, error) {
	r.log.Debug("Buscando campanha por ID", "id", campaignID)

	query := `
		SELECT id, account_id, name, description, channels, filters, status, created_at, updated_at
		FROM campaigns WHERE id = $1
	`
	campaign := &models.Campaign{}
	var channelsJSON, filtersJSON []byte

	err := r.db.QueryRow(query, campaignID).Scan(
		&campaign.ID, &campaign.AccountID, &campaign.Name, &campaign.Description,
		&channelsJSON, &filtersJSON, &campaign.Status, &campaign.CreatedAt, &campaign.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Campanha não encontrada
		}
		return nil, fmt.Errorf("erro ao buscar campanha: %w", err)
	}

	// Desserializar JSONB
	if err := json.Unmarshal(channelsJSON, &campaign.Channels); err != nil {
		r.log.Warn("Erro ao desserializar channels", "error", err)
	}
	if err := json.Unmarshal(filtersJSON, &campaign.Filters); err != nil {
		r.log.Warn("Erro ao desserializar filters", "error", err)
	}

	r.log.Debug("Campanha encontrada", "id", campaign.ID)
	return campaign, nil
}

// GetAllByAccountID retorna todas as campanhas de uma conta
func (r *campaignRepo) GetAllByAccountID(accountID uuid.UUID) ([]models.Campaign, error) {
	r.log.Debug("Buscando todas as campanhas da conta", "account_id", accountID)

	query := `
		SELECT id, account_id, name, description, channels, filters, status, created_at, updated_at
		FROM campaigns WHERE account_id = $1
	`
	rows, err := r.db.Query(query, accountID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar campanhas: %w", err)
	}
	defer rows.Close()

	var campaigns []models.Campaign
	for rows.Next() {
		var campaign models.Campaign
		var channelsJSON, filtersJSON []byte

		if err := rows.Scan(
			&campaign.ID, &campaign.AccountID, &campaign.Name, &campaign.Description,
			&channelsJSON, &filtersJSON, &campaign.Status, &campaign.CreatedAt, &campaign.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear campanhas: %w", err)
		}

		// Desserializa JSONB
		if err := json.Unmarshal(channelsJSON, &campaign.Channels); err != nil {
			r.log.Warn("Erro ao desserializar channels", "error", err)
		}
		if err := json.Unmarshal(filtersJSON, &campaign.Filters); err != nil {
			r.log.Warn("Erro ao desserializar filters", "error", err)
		}

		campaigns = append(campaigns, campaign)
	}

	r.log.Debug("Total de campanhas encontradas", "count", len(campaigns))

	return campaigns, nil
}

// UpdateByID atualiza os dados de uma campanha
func (r *campaignRepo) UpdateByID(campaignID uuid.UUID, campaign *models.Campaign) (*models.Campaign, error) {
	r.log.Debug("Atualizando campanha", "id", campaignID)

	channelsJSON, err := json.Marshal(campaign.Channels)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar channels: %w", err)
	}
	filtersJSON, err := json.Marshal(campaign.Filters)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar filters: %w", err)
	}

	query := `
		UPDATE campaigns
		SET name = $1, description = $2, channels = $3, filters = $4, status = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING updated_at
	`
	err = r.db.QueryRow(
		query, campaign.Name, campaign.Description, channelsJSON, filtersJSON, campaign.Status, campaignID,
	).Scan(&campaign.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar campanha: %w", err)
	}

	r.log.Debug("Campanha atualizada com sucesso", "id", campaignID)
	return campaign, nil
}

// UpdateStatus atualiza apenas o status da campanha
func (r *campaignRepo) UpdateStatus(campaignID uuid.UUID, status string) error {
	r.log.Debug("Atualizando status da campanha", "id", campaignID, "status", status)

	query := `UPDATE campaigns SET status = $1, updated_at = NOW()  WHERE id = $2`
	_, err := r.db.Exec(query, status, campaignID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar status da campanha: %w", err)
	}

	r.log.Debug("Status da campanha atualizado com sucesso", "id", campaignID)
	return nil
}

// DeleteByID remove uma campanha pelo ID
func (r *campaignRepo) DeleteByID(campaignID uuid.UUID) error {
	r.log.Debug("Deletando campanha", "id", campaignID)

	query := `DELETE FROM campaigns WHERE id = $1`
	_, err := r.db.Exec(query, campaignID)
	if err != nil {
		return fmt.Errorf("erro ao deletar campanha: %w", err)
	}

	r.log.Debug("Campanha deletada com sucesso", "id", campaignID)
	return nil
}
