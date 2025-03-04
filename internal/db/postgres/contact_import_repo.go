// File: /internal/postgres/contact_import_repo.go

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type contactImportRepo struct {
	log *slog.Logger
	db  *sql.DB
}

func NewContactImportRepository(db *sql.DB) db.ContactImportRepository {
	return &contactImportRepo{log: logger.GetLogger(), db: db}
}

func (r *contactImportRepo) Create(ctx context.Context, importData *models.ContactImport) (*models.ContactImport, error) {
	configJSON, _ := json.Marshal(importData.Config)
	previewJSON, _ := json.Marshal(importData.Preview)

	query := `
		INSERT INTO contact_imports (account_id, file_name, status, config, preview_data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		importData.AccountID, importData.FileName, importData.Status, configJSON, previewJSON,
	).Scan(&importData.ID, &importData.CreatedAt, &importData.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return importData, nil
}

func (r *contactImportRepo) GetAllByAccountID(ctx context.Context, accountID uuid.UUID) ([]models.ContactImport, error) {
	query := `
		SELECT id, file_name, status, created_at, updated_at
		FROM contact_imports
		WHERE account_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var imports []models.ContactImport
	for rows.Next() {
		var imp models.ContactImport
		err := rows.Scan(&imp.ID, &imp.FileName, &imp.Status, &imp.CreatedAt, &imp.UpdatedAt)
		if err != nil {
			return nil, err
		}
		imports = append(imports, imp)
	}
	return imports, nil
}

func (r *contactImportRepo) GetByID(ctx context.Context, accountID uuid.UUID, importID uuid.UUID) (*models.ContactImport, error) {
	query := `
		SELECT id, file_name, status, config, preview_data, created_at, updated_at
		FROM contact_imports
		WHERE account_id = $1 AND id = $2
	`
	row := r.db.QueryRowContext(ctx, query, accountID, importID)

	var imp models.ContactImport
	var configJSON, previewJSON []byte

	err := row.Scan(&imp.ID, &imp.FileName, &imp.Status, &configJSON, &previewJSON, &imp.CreatedAt, &imp.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Converte JSONB para Go Structs
	json.Unmarshal(configJSON, &imp.Config)
	json.Unmarshal(previewJSON, &imp.Preview)

	return &imp, nil
}

func (r *contactImportRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE contact_imports SET status = $1, updated_at = now() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *contactImportRepo) UpdateConfig(ctx context.Context, accountID uuid.UUID, id uuid.UUID, config models.ContactImportConfig) (*models.ContactImport, error) {
	// 🔹 Converte a estrutura para JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter configuração para JSON: %w", err)
	}

	query := `
		UPDATE contact_imports 
		SET config = $1, updated_at = now() 
		WHERE account_id = $2 AND id = $3
		RETURNING id, created_at, updated_at
	`

	var imp models.ContactImport
	// 🔹 Corrigida a ordem dos parâmetros na query
	err = r.db.QueryRowContext(ctx, query, configJSON, accountID, id).Scan(&imp.ID, &imp.CreatedAt, &imp.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("erro ao atualizar configuração no banco: %w", err)
	}

	return &imp, nil
}
