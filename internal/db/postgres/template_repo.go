// File: /internal/db/postgres/template_repo.go

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// templateRepository implementa TemplateRepositorysitory para PostgreSQL.
type templateRepository struct {
	log *slog.Logger
	db  *sql.DB
}

// NewTemplateRepository cria um novo repositório para templates.
func NewTemplateRepository(db *sql.DB) db.TemplateRepository {
	log := logger.GetLogger()
	return &templateRepository{log: log, db: db}
}

// Create insere um novo template no banco de dados.
func (r *templateRepository) Create(ctx context.Context, template *models.Template) (*models.Template, error) {
	r.log.Debug("Criando novo template", "template", template)
	query := `
		INSERT INTO templates (
			account_id, name, description, channel
		) VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRow(
		query,
		template.AccountID, template.Name, template.Description, template.Channel,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao criar template: %w", err)
	}

	r.log.Debug("Template criado com sucesso", "id", template.ID)

	return template, nil
}

// GetByID retorna um template pelo ID.
func (r *templateRepository) GetByID(ctx context.Context, templateID uuid.UUID) (*models.Template, error) {
	r.log.Debug("Buscando template por ID", "id", templateID)

	query := `
		SELECT id, account_id, name, description, channel, created_at, updated_at
		FROM templates WHERE id = $1
	`
	template := &models.Template{}
	err := r.db.QueryRow(query, templateID).Scan(
		&template.ID, &template.AccountID, &template.Name, &template.Description,
		&template.Channel, &template.CreatedAt, &template.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("erro ao buscar template: %w", err)
	}

	r.log.Debug("Template encontrado", "id", template.ID)

	return template, nil
}

// GetByAccountID retorna todos os templates de uma conta específica.
func (r *templateRepository) GetByAccountID(ctx context.Context, accountID uuid.UUID, filters map[string]string) ([]models.Template, error) {
	r.log.Debug("Buscando templates por account_id", "account_id", accountID, "filters", filters)

	baseQuery := `
		SELECT id, account_id, name, description, channel, created_at, updated_at
		FROM templates
		WHERE account_id = $1
	`
	args := []interface{}{accountID}
	filterIndex := 2

	// 🔍 Aplicando filtros dinâmicos
	for key, value := range filters {
		switch key {
		case "name":
			baseQuery += fmt.Sprintf(" AND name ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
		case "description":
			baseQuery += fmt.Sprintf(" AND description ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
		case "channel":
			baseQuery += fmt.Sprintf(" AND channel = $%d", filterIndex)
			args = append(args, value)
		}
		filterIndex++
	}

	baseQuery += " ORDER BY created_at DESC"

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar templates: %w", err)
	}
	defer rows.Close()

	var templates []models.Template
	for rows.Next() {
		var template models.Template
		if err := rows.Scan(
			&template.ID, &template.AccountID, &template.Name, &template.Description, &template.Channel, &template.CreatedAt, &template.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear templates: %w", err)
		}
		templates = append(templates, template)
	}

	r.log.Debug("Templates encontrados", "total", len(templates))

	return templates, nil
}

// UpdateByID atualiza um template pelo ID.
func (r *templateRepository) UpdateByID(ctx context.Context, templateID uuid.UUID, template *models.Template) (*models.Template, error) {
	r.log.Debug("Atualizando template por ID", "id", templateID)

	query := `
		UPDATE templates
		SET name = $1, description = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING updated_at
	`
	err := r.db.QueryRow(
		query,
		template.Name, template.Description, templateID,
	).Scan(&template.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar template: %w", err)
	}

	r.log.Debug("Template atualizado com sucesso", "id", templateID)

	return template, nil
}

// DeleteByID remove um template pelo ID.
func (r *templateRepository) DeleteByID(ctx context.Context, templateID uuid.UUID) error {
	r.log.Debug("Deletando template por ID", "id", templateID)

	query := `DELETE FROM templates WHERE id = $1`
	_, err := r.db.Exec(query, templateID)
	if err != nil {
		return fmt.Errorf("erro ao deletar template: %w", err)
	}

	r.log.Debug("Template deletado com sucesso", "id", templateID)

	return nil
}
