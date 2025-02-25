// File: /internal/db/postgres/contact_repo.go

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// contactRepo implementa ContactRepository para PostgreSQL.
type contactRepo struct {
	log *slog.Logger
	db  *sql.DB
}

// NewContactRepository cria um novo repositório para contatos.
func NewContactRepository(db *sql.DB) *contactRepo {
	log := logger.GetLogger()
	return &contactRepo{log: log, db: db}
}

// 📌 Criar contato no banco
func (r *contactRepo) Create(ctx context.Context, contact *models.Contact) (*models.Contact, error) {
	r.log.Debug("Criando novo contato",
		slog.String("account_id", contact.AccountID.String()),
		slog.String("name", contact.Name),
		slog.String("email", utils.SafeString(contact.Email)),
		slog.String("whatsapp", utils.SafeString(contact.WhatsApp)))

	tagsJSON, err := json.Marshal(contact.Tags)
	if err != nil {
		return nil, fmt.Errorf("erro ao converter tags para JSON: %w", err)
	}

	query := `
		INSERT INTO contacts (
			account_id, name, email, whatsapp, gender, birth_date, bairro, cidade, estado, tags, history, opt_out_at, last_contact_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`

	err = r.db.QueryRow(
		query,
		contact.AccountID, contact.Name, contact.Email, contact.WhatsApp,
		contact.Gender, contact.BirthDate, contact.Bairro, contact.Cidade, contact.Estado,
		tagsJSON, contact.History, contact.OptOutAt, contact.LastContactAt,
	).Scan(&contact.ID, &contact.CreatedAt, &contact.UpdatedAt)

	if err != nil {
		r.log.Error("Erro ao criar contato",
			slog.String("account_id", contact.AccountID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao criar contato: %w", err)
	}

	r.log.Info("Contato criado com sucesso",
		slog.String("contact_id", contact.ID.String()),
		slog.String("account_id", contact.AccountID.String()))

	return contact, nil
}

// 📌 Buscar contato pelo ID
func (r *contactRepo) GetByID(ctx context.Context, contactID uuid.UUID) (*models.Contact, error) {
	r.log.Debug("Buscando contato por ID",
		slog.String("contact_id", contactID.String()))

	query := `
		SELECT id, account_id, name, email, whatsapp, gender, birth_date, bairro, cidade, estado, tags, history, opt_out_at, last_contact_at, created_at, updated_at
		FROM contacts WHERE id = $1
	`

	contact := &models.Contact{}
	var tagsJSON []byte

	err := r.db.QueryRow(query, contactID).Scan(
		&contact.ID, &contact.AccountID, &contact.Name, &contact.Email, &contact.WhatsApp,
		&contact.Gender, &contact.BirthDate, &contact.Bairro, &contact.Cidade, &contact.Estado,
		&tagsJSON, &contact.History, &contact.OptOutAt, &contact.LastContactAt,
		&contact.CreatedAt, &contact.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		r.log.Error("Erro ao buscar contato",
			slog.String("contact_id", contactID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao buscar contato: %w", err)
	}

	if len(tagsJSON) > 0 {
		_ = json.Unmarshal(tagsJSON, &contact.Tags)
	}

	r.log.Debug("Contato encontrado",
		slog.String("contact_id", contact.ID.String()),
		slog.String("account_id", contact.AccountID.String()))

	return contact, nil
}

// 📌 Buscar contatos por account_id
func (r *contactRepo) GetByAccountID(ctx context.Context, accountID uuid.UUID, filters map[string]string) ([]models.Contact, error) {
	r.log.Debug("Buscando contatos por account_id",
		slog.String("account_id", accountID.String()))

	baseQuery := `
		SELECT id, account_id, name, email, whatsapp, gender, birth_date, bairro, cidade, estado, tags, history, opt_out_at, last_contact_at, created_at, updated_at
		FROM contacts
		WHERE account_id = $1
	`
	args := []interface{}{accountID}
	filterIndex := 2

	// 🔍 Aplicando filtros dinâmicos
	for key, value := range filters {
		switch key {
		case "name", "email", "whatsapp", "cidade", "estado", "bairro":
			baseQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, filterIndex)
			args = append(args, "%"+value+"%")
		case "tags":
			// Busca qualquer tag que contenha o valor passado
			baseQuery += fmt.Sprintf(" AND tags::text ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
		case "interesses":
			// Busca dentro da chave "interesses" do JSONB
			baseQuery += fmt.Sprintf(" AND tags->'interesses' ? $%d", filterIndex)
			args = append(args, value)
		case "perfil":
			// Busca dentro da chave "perfil" do JSONB
			baseQuery += fmt.Sprintf(" AND tags->'perfil' ? $%d", filterIndex)
			args = append(args, value)
		case "eventos":
			// Busca dentro da chave "eventos" do JSONB
			baseQuery += fmt.Sprintf(" AND tags->'eventos' ? $%d", filterIndex)
			args = append(args, value)
		}
		filterIndex++
	}

	baseQuery += " ORDER BY created_at DESC"

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar contatos: %w", err)
	}
	defer rows.Close()

	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		var tagsJSON []byte

		if err := rows.Scan(
			&contact.ID, &contact.AccountID, &contact.Name, &contact.Email, &contact.WhatsApp,
			&contact.Gender, &contact.BirthDate, &contact.Bairro, &contact.Cidade, &contact.Estado,
			&tagsJSON, &contact.History, &contact.OptOutAt, &contact.LastContactAt,
			&contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear contatos: %w", err)
		}

		// Decodificar JSONB para Tags
		_ = json.Unmarshal(tagsJSON, &contact.Tags)

		contacts = append(contacts, contact)
	}

	r.log.Debug("Contatos encontrados",
		slog.String("account_id", accountID.String()),
		slog.Int("total", len(contacts)))
	return contacts, nil
}

// 📌 Buscar contato por e-mail ou WhatsApp dentro de uma conta específica
func (r *contactRepo) FindByEmailOrWhatsApp(ctx context.Context, accountID uuid.UUID, email, whatsapp *string) (*models.Contact, error) {
	query := `
		SELECT id, account_id, name, email, whatsapp FROM contacts
		WHERE account_id = $1 AND (email = $2 OR whatsapp = $3)
		LIMIT 1
	`

	contact := &models.Contact{}

	err := r.db.QueryRow(query, accountID, email, whatsapp).Scan(
		&contact.ID, &contact.AccountID, &contact.Name, &contact.Email, &contact.WhatsApp,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Nenhum contato encontrado
		}
		r.log.Error("Erro ao buscar contato por email/WhatsApp",
			slog.String("account_id", accountID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao buscar contato: %w", err)
	}

	r.log.Debug("Contato encontrado por email/WhatsApp",
		slog.String("contact_id", contact.ID.String()),
		slog.String("account_id", accountID.String()))

	return contact, nil
}

// 📌 Atualizar contato pelo ID
func (r *contactRepo) UpdateByID(ctx context.Context, contactID uuid.UUID, contact *models.Contact) (*models.Contact, error) {
	r.log.Debug("Atualizando contato",
		slog.String("contact_id", contactID.String()),
		slog.String("account_id", contact.AccountID.String()))

	tagsJSON, _ := json.Marshal(contact.Tags)

	query := `
		UPDATE contacts
		SET name = $1, email = $2, whatsapp = $3, gender = $4, birth_date = $5, bairro = $6, cidade = $7, estado = $8, tags = $9, history = $10, opt_out_at = $11, updated_at = NOW()
		WHERE id = $12
		RETURNING updated_at
	`
	err := r.db.QueryRow(
		query,
		contact.Name, contact.Email, contact.WhatsApp, contact.Gender, contact.BirthDate,
		contact.Bairro, contact.Cidade, contact.Estado, tagsJSON, contact.History, contact.OptOutAt, contactID,
	).Scan(&contact.UpdatedAt)

	if err != nil {
		r.log.Error("Erro ao atualizar contato",
			slog.String("contact_id", contactID.String()),
			slog.String("account_id", contact.AccountID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao atualizar contato: %w", err)
	}

	r.log.Info("Contato atualizado com sucesso",
		slog.String("contact_id", contactID.String()),
		slog.String("account_id", contact.AccountID.String()))

	return contact, nil
}

// 📌 Deletar contato pelo ID
func (r *contactRepo) DeleteByID(ctx context.Context, contactID uuid.UUID) error {
	r.log.Debug("Deletando contato",
		slog.String("contact_id", contactID.String()))

	query := `DELETE FROM contacts WHERE id = $1`
	_, err := r.db.Exec(query, contactID)
	if err != nil {
		r.log.Error("Erro ao deletar contato",
			slog.String("contact_id", contactID.String()),
			slog.String("error", err.Error()))
		return fmt.Errorf("erro ao deletar contato: %w", err)
	}

	r.log.Info("Contato deletado com sucesso",
		slog.String("contact_id", contactID.String()))

	return nil
}
