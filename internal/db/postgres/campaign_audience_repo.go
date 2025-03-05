// File: /internal/postgres/campaign_audience_repo.go

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
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
func (r *campaignAudienceRepo) AddContactsToCampaign(ctx context.Context, campaignID uuid.UUID, contacts []models.CampaignAudience) ([]models.CampaignAudience, error) {
	audiences := []models.CampaignAudience{}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	stmt, err := tx.Prepare(`
		INSERT INTO campaigns_audience (campaign_id, contact_id, type, status, updated_at)
		VALUES ($1, $2, $3, 'pendente', NOW())
		ON CONFLICT (campaign_id, contact_id) DO UPDATE 
		SET updated_at = NOW()
		RETURNING id, campaign_id, contact_id, type, status, message_id, feedback_api, created_at, updated_at
	`)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	defer stmt.Close()

	for _, contact := range contacts {
		var audience models.CampaignAudience
		err := stmt.QueryRow(campaignID, contact.ContactID, contact.Type).Scan(
			&audience.ID, &audience.CampaignID, &audience.ContactID, &audience.Type, &audience.Status, &audience.MessageID, &audience.Feedback, &audience.CreatedAt, &audience.UpdatedAt,
		)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		audiences = append(audiences, audience)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return audiences, nil
}

// Adiciona todos os contatos filtrados à audiência da campanha
func (r *campaignAudienceRepo) AddAllFilteredContacts(ctx context.Context, accountID uuid.UUID, campaignID uuid.UUID, filters *map[string]string, channelType models.ChannelType) error {
	// 🔍 Query base para buscar contatos filtrados
	selectQuery := `
		SELECT id
		FROM contacts
		WHERE account_id = $1 
		AND opt_out_at IS NULL 
		AND id NOT IN (SELECT contact_id FROM campaigns_audience WHERE campaign_id = $2)
	`

	if channelType == models.EmailChannel {
		selectQuery += " AND email IS NOT NULL"
	} else if channelType == models.WhatsappChannel {
		selectQuery += " AND whatsapp IS NOT NULL"
	}

	args := []interface{}{accountID, campaignID}
	filterIndex := 3

	// 🔍 Aplicar filtros dinâmicos
	if filters != nil && len(*filters) > 0 {
		for key, value := range *filters {
			switch key {
			case "name", "email", "whatsapp", "cidade", "estado", "bairro":
				selectQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, filterIndex)
				args = append(args, "%"+value+"%")
			case "gender":
				selectQuery += fmt.Sprintf(" AND gender = $%d", filterIndex)
				args = append(args, value)
			case "birth_date_start":
				selectQuery += fmt.Sprintf(" AND birth_date >= $%d", filterIndex)
				dateValue, err := utils.ParseDate(value)
				if err != nil {
					return fmt.Errorf("erro ao converter data: %w", err)
				}
				args = append(args, dateValue)
			case "birth_date_end":
				selectQuery += fmt.Sprintf(" AND birth_date <= $%d", filterIndex)
				dateValue, err := utils.ParseDate(value)
				if err != nil {
					return fmt.Errorf("erro ao converter data: %w", err)
				}
				args = append(args, dateValue)
			case "last_contact_at":
				selectQuery += fmt.Sprintf(" AND last_contact_at >= $%d", filterIndex)
				dateValue, err := utils.ParseDate(value)
				if err != nil {
					return fmt.Errorf("erro ao converter data: %w", err)
				}
				args = append(args, dateValue)
			case "tags":
				// Filtra globalmente dentro do JSONB como um texto simples
				tags := strings.Split(value, ",")
				var conditions []string

				for _, tag := range tags {
					tag = strings.TrimSpace(tag)
					if tag == "" {
						continue
					}
					conditions = append(conditions, fmt.Sprintf("tags::TEXT ILIKE $%d", filterIndex))
					args = append(args, "%"+tag+"%")
					filterIndex++
				}

				if len(conditions) > 0 {
					selectQuery += " AND (" + strings.Join(conditions, " OR ") + ")"
				}
			case "interesses", "perfil", "eventos":
				// Filtra dentro de uma subcategoria específica do JSONB
				tags := strings.Split(value, ",")
				var conditions []string

				for _, tag := range tags {
					tag = strings.TrimSpace(tag)
					if tag == "" {
						continue
					}
					conditions = append(conditions, fmt.Sprintf(
						"EXISTS (SELECT 1 FROM jsonb_array_elements_text(tags->'%s') AS t WHERE LOWER(t) ILIKE $%d)",
						key, filterIndex,
					))
					args = append(args, "%"+strings.ToLower(tag)+"%")
					filterIndex++
				}

				if len(conditions) > 0 {
					selectQuery += " AND (" + strings.Join(conditions, " OR ") + ")"
				}

			}
			filterIndex++
		}
	}

	// 🔹 Buscar IDs dos contatos disponíveis
	rows, err := r.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return fmt.Errorf("erro ao buscar contatos disponíveis: %w", err)
	}
	defer rows.Close()

	// 🔍 Criar inserção em lote
	var placeholders []string
	insertArgs := []interface{}{campaignID} // Primeiro argumento é o ID da campanha
	insertIndex := 2

	for rows.Next() {
		var contactID uuid.UUID
		if err := rows.Scan(&contactID); err != nil {
			return fmt.Errorf("erro ao escanear contatos disponíveis: %w", err)
		}

		placeholders = append(placeholders, fmt.Sprintf("($1, $%d, $%d)", insertIndex, insertIndex+1))
		insertArgs = append(insertArgs, contactID, channelType)
		insertIndex += 2
	}

	// 🔹 Executar INSERT em lote
	if len(placeholders) > 0 {
		insertQuery := fmt.Sprintf(`
			INSERT INTO campaigns_audience (campaign_id, contact_id, type) 
			VALUES %s
			ON CONFLICT (campaign_id, contact_id) DO NOTHING`, strings.Join(placeholders, ","),
		)

		if _, err := r.db.ExecContext(ctx, insertQuery, insertArgs...); err != nil {
			return fmt.Errorf("erro ao inserir contatos filtrados: %w", err)
		}
	}

	return nil
}

// GetCampaignAudience retorna a audiência de uma campanha junto com os detalhes dos contatos
func (r *campaignAudienceRepo) GetCampaignAudience(ctx context.Context, campaignID uuid.UUID, contactType *string) ([]dto.CampaignAudienceDTO, error) {
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

	var messages []dto.CampaignAudienceDTO
	for rows.Next() {
		var msg dto.CampaignAudienceDTO
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
func (r *campaignAudienceRepo) RemoveContactFromCampaign(ctx context.Context, campaignID, audienceID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM campaigns_audience WHERE campaign_id = $1 AND id = $2`, campaignID, audienceID)
	return err
}

// Atualiza o status de um contato enviado
func (r *campaignAudienceRepo) UpdateStatus(ctx context.Context, contactID uuid.UUID, status, messageID string, feedback map[string]interface{}) error {
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
func (r *campaignAudienceRepo) UpdateStatusByMessageID(ctx context.Context, messageID string, status string, feedbackAPI *string) error {
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

// GetCampaignAudienceToSQS busca a audiência da campanha para envio à fila SQS.
func (r *campaignAudienceRepo) GetCampaignAudienceToSQS(ctx context.Context, accountID uuid.UUID, campaignID uuid.UUID, contactType *string) ([]dto.CampaignMessageDTO, error) {
	// Query base para buscar os contatos da campanha
	query := `
		SELECT
			ca.id,
			ca.campaign_id,
			ca.contact_id,
			ca.type
		FROM
			campaigns_audience ca
		WHERE
			ca.campaign_id = $1
	`

	args := []interface{}{campaignID} // ✅ Correção: Passa UUID diretamente

	// Adiciona filtro opcional pelo tipo de contato (email ou WhatsApp)
	if contactType != nil {
		query += " AND ca.type = $2"
		args = append(args, *contactType)
	}

	r.log.Debug("🔍 Buscando audiência da campanha", slog.String("query", query))
	r.log.Debug("🔍 Buscando audiência da campanha", slog.String("args", fmt.Sprintf("%v", args)))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []dto.CampaignMessageDTO
	for rows.Next() {
		msg := dto.CampaignMessageDTO{AccountID: accountID}
		if err := rows.Scan(&msg.ID, &msg.CampaignID, &msg.ContactID, &msg.Type); err != nil {
			return nil, err
		}

		messages = append(messages, msg)
	}

	r.log.Debug("✅ Audiência da campanha encontrada",
		slog.String("account_id", accountID.String()),
		slog.String("campaign_id", campaignID.String()),
		slog.Int("total", len(messages)),
	)

	return messages, nil
}

// GetPaginatedCampaignAudience retorna a audiência de uma campanha com paginação
func (r *campaignAudienceRepo) GetPaginatedCampaignAudience(
	ctx context.Context,
	campaignID uuid.UUID,
	contactType *string,
	currentPage int,
	perPage int,
) (*models.Paginator, error) {
	// Garantir valores mínimos para paginação
	if currentPage < 1 {
		currentPage = 1
	}
	if perPage < 1 {
		perPage = 10
	}

	// Query base
	baseQuery := `
		SELECT ca.id, ca.campaign_id, ca.contact_id, ca.type, ca.status,
		       c.name, c.email, c.whatsapp, c.gender, c.birth_date, 
		       c.bairro, c.cidade, c.estado, c.tags, c.history, 
		       c.opt_out_at, c.last_contact_at, c.created_at, c.updated_at
		FROM campaigns_audience ca
		INNER JOIN contacts c ON ca.contact_id = c.id
		WHERE ca.campaign_id = $1
	`

	args := []interface{}{campaignID}
	filterIndex := 2

	if contactType != nil {
		baseQuery += " AND ca.type = $2"
		args = append(args, *contactType)
		filterIndex++
	}

	// Contar total de registros antes da paginação
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS total"
	var totalRecords int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		return nil, fmt.Errorf("erro ao contar audiência da campanha: %w", err)
	}

	// Calcular total de páginas
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Aplicar ordenação
	baseQuery += " ORDER BY c.updated_at DESC"

	// Aplicar paginação
	offset := (currentPage - 1) * perPage
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)

	// Executar query de busca
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar audiência da campanha: %w", err)
	}
	defer rows.Close()

	// Preencher os resultados
	var audience []dto.CampaignAudienceDTO
	for rows.Next() {
		var msg dto.CampaignAudienceDTO
		var tagsJSON []byte

		if err := rows.Scan(
			&msg.ID, &msg.CampaignID, &msg.ContactID, &msg.Type, &msg.Status,
			&msg.Name, &msg.Email, &msg.WhatsApp, &msg.Gender, &msg.BirthDate,
			&msg.Bairro, &msg.Cidade, &msg.Estado, &tagsJSON, &msg.History,
			&msg.OptOutAt, &msg.LastContactAt, &msg.CreatedAt, &msg.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear audiência da campanha: %w", err)
		}

		// Converter JSONB das tags para map[string]interface{}
		if err := json.Unmarshal(tagsJSON, &msg.Tags); err != nil {
			msg.Tags = make(map[string]interface{}) // Se der erro, inicializa vazio
		}

		audience = append(audience, msg)
	}

	// Retornar página de resultados
	return &models.Paginator{
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		PerPage:      perPage,
		Data:         audience,
	}, nil
}

// RemoveAllContactsFromCampaign remove todos os contatos de uma campanha
func (r *campaignAudienceRepo) RemoveAllContactsFromCampaign(ctx context.Context, campaignID uuid.UUID) error {
	_, err := r.db.Exec(`DELETE FROM campaigns_audience WHERE campaign_id = $1`, campaignID)
	return err
}
