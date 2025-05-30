// internal/db/postgres/whatsapp_contact_repo.go

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

// whatsappContactRepo implementa ContactRepository para PostgreSQL.
type whatsappContactRepo struct {
	log *slog.Logger
	db  *sql.DB
}

// NewWhatsappContactRepository cria um novo repositório para contatos.
func NewWhatsappContactRepository(db *sql.DB) db.WhatsappContactRepository {
	log := logger.GetLogger()
	return &whatsappContactRepo{log: log, db: db}
}

// FindOrCreate insere ou atualiza um contato do WhatsApp no banco de dados.
func (r *whatsappContactRepo) FindOrCreate(ctx context.Context, contact *models.WhatsappContact) (*models.WhatsappContact, error) {
	// Serializa o campo BusinessProfile (pode ser nil)
	var businessJSON []byte
	if contact.BusinessProfile != nil {
		var err error
		businessJSON, err = json.Marshal(contact.BusinessProfile)
		if err != nil {
			return nil, fmt.Errorf("erro ao serializar business_profile: %w", err)
		}
	} else {
		businessJSON = []byte("null") // ✅ importante para compatibilidade com JSONB
	}

	query := `
		INSERT INTO whatsapp_contacts (
			account_id, contact_id, name, phone, jid, is_business, business_profile
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		)
		ON CONFLICT (account_id, jid) DO UPDATE
		SET updated_at = now()
		RETURNING id, account_id, contact_id, name, phone, jid, is_business, business_profile, created_at, updated_at;
	`

	row := r.db.QueryRowContext(ctx, query,
		contact.AccountID,
		contact.ContactID,
		contact.Name,
		contact.Phone,
		contact.JID,
		contact.IsBusiness,
		businessJSON,
	)

	var result models.WhatsappContact
	var businessRaw []byte

	err := row.Scan(
		&result.ID,
		&result.AccountID,
		&result.ContactID,
		&result.Name,
		&result.Phone,
		&result.JID,
		&result.IsBusiness,
		&businessRaw,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("erro ao inserir ou recuperar contato do WhatsApp: %w", err)
	}

	// Deserializa o business_profile se presente
	if len(businessRaw) > 0 && string(businessRaw) != "null" {
		var bd models.BusinessProfile
		if err := json.Unmarshal(businessRaw, &bd); err != nil {
			log.Printf("⚠️ erro ao desserializar business_profile: %v", err)
		} else {
			result.BusinessProfile = &bd
		}
	}

	return &result, nil
}

// FindByID busca um contato do WhatsApp pelo ID.
func (r *whatsappContactRepo) FindByID(ctx context.Context, id uuid.UUID) (*models.WhatsappContact, error) {
	query := `
		SELECT id, account_id, contact_id, name, phone, jid, is_business, business_profile, created_at, updated_at
		FROM whatsapp_contacts
		WHERE id = $1
		LIMIT 1;
	`

	r.log.Debug("SQL Query", "query", query, "id", id)

	row := r.db.QueryRowContext(ctx, query, id)

	var result models.WhatsappContact
	var businessRaw []byte

	err := row.Scan(
		&result.ID,
		&result.AccountID,
		&result.ContactID,
		&result.Name,
		&result.Phone,
		&result.JID,
		&result.IsBusiness,
		&businessRaw,
		&result.CreatedAt,
		&result.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Contato não encontrado
		}
		return nil, fmt.Errorf("erro ao buscar contato do WhatsApp: %w", err)
	}

	if len(businessRaw) > 0 && string(businessRaw) != "null" {
		var bd models.BusinessProfile
		if err := json.Unmarshal(businessRaw, &bd); err != nil {
			r.log.Error("erro ao desserializar business_profile", "error", err)
			return &result, nil // Retorna mesmo com erro de deserialização
		}
		result.BusinessProfile = &bd
	}

	return &result, nil
}
