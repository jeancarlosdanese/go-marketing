// internal/db/postgres/chat_contact_repo.go

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type chatContactRepository struct {
	log slog.Logger
	db  *sql.DB
}

func NewChatContactRepository(db *sql.DB) db.ChatContactRepository {
	return &chatContactRepository{log: *logger.GetLogger(), db: db}
}

// ✅ Busca atendimento existente ou cria um novo para o contato no chat
func (r *chatContactRepository) FindOrCreate(ctx context.Context, accountID, chatID, whatsappContactID uuid.UUID) (*models.ChatContact, error) {
	// Primeiro tenta encontrar
	query := `
		SELECT id, account_id, chat_id, whatsapp_contact_id, status, created_at, updated_at
		FROM chat_contacts
		WHERE account_id = $1 AND chat_id = $2 AND whatsapp_contact_id = $3
		LIMIT 1
	`

	var chatContact models.ChatContact
	err := r.db.QueryRowContext(ctx, query, accountID, chatID, whatsappContactID).Scan(
		&chatContact.ID,
		&chatContact.AccountID,
		&chatContact.ChatID,
		&chatContact.WhatsappContactID,
		&chatContact.Status,
		&chatContact.CreatedAt,
		&chatContact.UpdatedAt,
	)
	if err == nil {
		return &chatContact, nil
	}

	// Se não encontrou, cria novo
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	newContact := &models.ChatContact{
		AccountID:         accountID,
		ChatID:            chatID,
		WhatsappContactID: whatsappContactID,
		Status:            "aberto",
	}

	insertQuery := `
		INSERT INTO chat_contacts (
			account_id, chat_id, whatsapp_contact_id, status
		) VALUES ($1, $2, $3, $4)
		RETURNING id, account_id, chat_id, whatsapp_contact_id, status, created_at, updated_at
	`

	err = r.db.QueryRowContext(ctx, insertQuery,
		newContact.AccountID,
		newContact.ChatID,
		newContact.WhatsappContactID,
		newContact.Status,
	).Scan(
		&newContact.ID,
		&newContact.AccountID,
		&newContact.ChatID,
		&newContact.WhatsappContactID,
		&newContact.Status,
		&newContact.CreatedAt,
		&newContact.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return newContact, nil
}

// FindByID busca um contato de chat pelo ID
func (r *chatContactRepository) FindByID(ctx context.Context, accountID, chatID, chatContactID uuid.UUID) (*models.ChatContact, error) {
	query := `
		SELECT id, account_id, chat_id, whatsapp_contact_id, status, created_at, updated_at
		FROM chat_contacts
		WHERE account_id = $1 AND chat_id = $2 AND id = $3
		LIMIT 1
	`

	var chatContact models.ChatContact
	err := r.db.QueryRowContext(ctx, query, accountID, chatID, chatContactID).Scan(
		&chatContact.ID,
		&chatContact.AccountID,
		&chatContact.ChatID,
		&chatContact.WhatsappContactID,
		&chatContact.Status,
		&chatContact.CreatedAt,
		&chatContact.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("contato de chat não encontrado: %w", err)
		}
		return nil, err
	}

	return &chatContact, nil
}

// ListByChatID retorna todos os contatos de um chat
func (r *chatContactRepository) ListByChatID(ctx context.Context, accountID, chatID uuid.UUID) ([]dto.ChatContactFull, error) {
	query := `
		SELECT
			cc.id AS id,
			cc.chat_id AS chat_id,
			wc.contact_id AS contact_id,
			wc.id AS whatsapp_contact_id,
			wc.name AS name,
			wc.phone AS phone,
			wc.jid AS jid,
			wc.is_business AS is_business,
			cc.status AS status,
			cc.updated_at AS updated_at
		FROM
			chat_contacts cc
			INNER JOIN whatsapp_contacts wc ON wc.id = cc.whatsapp_contact_id
		WHERE
			cc.account_id = $1
			AND cc.chat_id = $2
		ORDER BY
			cc.updated_at DESC;
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, chatID)
	if err != nil {
		r.log.Error("Erro ao listar contatos do chat", slog.Any("err", err), slog.String("chat_id", chatID.String()))
		return nil, fmt.Errorf("erro ao listar contatos do chat: %w", err)
	}
	defer rows.Close()
	var chatContacts []dto.ChatContactFull
	for rows.Next() {
		var contact dto.ChatContactFull
		if err := rows.Scan(
			&contact.ID,
			&contact.ChatID,
			&contact.ContactID,
			&contact.WhatsappContactID,
			&contact.Name,
			&contact.Phone,
			&contact.JID,
			&contact.IsBusiness,
			&contact.Status,
			&contact.UpdatedAt,
		); err != nil {
			r.log.Error("Erro ao escanear contato do chat", slog.Any("err", err), slog.String("chat_id", chatID.String()))
			return nil, fmt.Errorf("erro ao escanear contato do chat: %w", err)
		}
		chatContacts = append(chatContacts, contact)
	}
	if err := rows.Err(); err != nil {
		r.log.Error("Erro ao iterar sobre contatos do chat", slog.Any("err", err), slog.String("chat_id", chatID.String()))
		return nil, fmt.Errorf("erro ao iterar sobre contatos do chat: %w", err)
	}
	r.log.Debug("Contatos do chat listados com sucesso", slog.Any("chat_id", chatID.String()), slog.Int("count", len(chatContacts)))

	return chatContacts, nil
}
