// File: /internal/db/postgres/contact_repo.go

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
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
func NewContactRepository(db *sql.DB) db.ContactRepository {
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

// GetPaginatedContacts busca contatos com paginação e filtros dinâmicos
func (r *contactRepo) GetPaginatedContacts(
	ctx context.Context,
	accountID uuid.UUID,
	filters map[string]string,
	sort string,
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
		SELECT id, name, email, whatsapp, gender, birth_date, bairro, cidade, estado, last_contact_at, created_at, updated_at
		FROM contacts
		WHERE account_id = $1 AND opt_out_at IS NULL
	`

	args := []interface{}{accountID}
	filterIndex := 2

	// Aplicar filtros dinâmicos
	for key, value := range filters {
		switch key {
		case "name", "email", "whatsapp", "cidade", "estado", "bairro":
			baseQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, filterIndex)
			args = append(args, "%"+value+"%")
			filterIndex++
		case "gender":
			baseQuery += fmt.Sprintf(" AND gender = $%d", filterIndex)
			args = append(args, value)
			filterIndex++
		case "birth_date_start":
			baseQuery += fmt.Sprintf(" AND birth_date >= $%d", filterIndex)
			start_date, err := utils.ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("erro ao converter data de nascimento: %w", err)
			}
			args = append(args, start_date)
			filterIndex++
		case "birth_date_end":
			baseQuery += fmt.Sprintf(" AND birth_date <= $%d", filterIndex)
			end_date, err := utils.ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("erro ao converter data de nascimento: %w", err)
			}
			args = append(args, end_date)
			filterIndex++
		case "last_contact_at":
			baseQuery += fmt.Sprintf(" AND last_contact_at >= $%d", filterIndex)
			last_contact_at, err := utils.ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("erro ao converter data de último contato: %w", err)
			}
			args = append(args, last_contact_at)
			filterIndex++
		case "interesses":
			// Busca parcial dentro da chave "interesses"
			baseQuery += fmt.Sprintf(" AND tags->>'interesses' ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
			filterIndex++
		case "perfil":
			// Busca parcial dentro da chave "perfil"
			baseQuery += fmt.Sprintf(" AND tags->>'perfil' ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
			filterIndex++
		case "eventos":
			// Busca parcial dentro da chave "eventos"
			baseQuery += fmt.Sprintf(" AND tags->>'eventos' ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
			filterIndex++
		case "tags":
			// Busca qualquer tag que contenha o valor passado
			baseQuery += fmt.Sprintf(" AND tags::text ILIKE $%d", filterIndex)
			args = append(args, "%"+value+"%")
			filterIndex++
		}
	}

	// Contar total de registros antes da paginação
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS total"
	var totalRecords int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		r.log.Debug("Erro ao contar contatos",
			slog.String("account_id", accountID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao contar contatos: %w", err)
	}

	// Calcular total de páginas
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Aplicar ordenação
	if sort != "" {
		baseQuery += fmt.Sprintf(" ORDER BY %s", sort)
	} else {
		baseQuery += " ORDER BY updated_at DESC"
	}

	r.log.Debug("Query de busca de contatos",
		slog.String("query", baseQuery))

	// Aplicar paginação
	offset := (currentPage - 1) * perPage
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)

	// Executar query de busca
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar contatos: %w", err)
	}
	defer rows.Close()

	// Preencher os resultados
	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact

		if err := rows.Scan(
			&contact.ID, &contact.Name, &contact.Email, &contact.WhatsApp, &contact.Gender,
			&contact.BirthDate, &contact.Bairro, &contact.Cidade, &contact.Estado,
			&contact.LastContactAt, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear contatos: %w", err)
		}

		contacts = append(contacts, contact)
	}

	// Retornar página de resultados
	return &models.Paginator{
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		PerPage:      perPage,
		Data:         contacts,
	}, nil
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

// ✅ **Buscar contatos disponíveis para uma campanha**
// GetAvailableContactsForCampaign busca contatos com paginação e filtros, excluindo os que já estão na campanha
func (r *contactRepo) GetAvailableContactsForCampaign(
	ctx context.Context,
	accountID uuid.UUID,
	campaignID uuid.UUID,
	filters map[string]string,
	sort string,
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

	// Query base (excluindo contatos que já estão na audiência da campanha)
	baseQuery := `
		SELECT id, name, email, whatsapp, gender, birth_date, bairro, cidade, estado, tags, last_contact_at, created_at, updated_at
		FROM contacts
		WHERE account_id = $1 
		AND opt_out_at IS NULL 
		AND id NOT IN (SELECT contact_id FROM campaigns_audience WHERE campaign_id = $2)
	`

	args := []interface{}{accountID, campaignID}
	filterIndex := 3

	// Aplicar filtros dinâmicos (mantendo a mesma lógica do método original)
	for key, value := range filters {
		switch key {
		case "name", "email", "whatsapp", "cidade", "estado", "bairro":
			baseQuery += fmt.Sprintf(" AND %s ILIKE $%d", key, filterIndex)
			args = append(args, "%"+value+"%")
		case "gender":
			baseQuery += fmt.Sprintf(" AND gender = $%d", filterIndex)
			args = append(args, value)
		case "birth_date_start":
			baseQuery += fmt.Sprintf(" AND birth_date >= $%d", filterIndex)
			startDate, err := utils.ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("erro ao converter data de nascimento: %w", err)
			}
			args = append(args, startDate)
		case "birth_date_end":
			baseQuery += fmt.Sprintf(" AND birth_date <= $%d", filterIndex)
			endDate, err := utils.ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("erro ao converter data de nascimento: %w", err)
			}
			args = append(args, endDate)
		case "last_contact_at":
			baseQuery += fmt.Sprintf(" AND last_contact_at >= $%d", filterIndex)
			lastContactAt, err := utils.ParseDate(value)
			if err != nil {
				return nil, fmt.Errorf("erro ao converter data de último contato: %w", err)
			}
			args = append(args, lastContactAt)
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
				baseQuery += " AND (" + strings.Join(conditions, " OR ") + ")"
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
				baseQuery += " AND (" + strings.Join(conditions, " OR ") + ")"
			}

		}
		filterIndex++
	}

	// Contar total de registros antes da paginação
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") AS total"
	var totalRecords int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalRecords); err != nil {
		r.log.Error("Erro ao contar contatos disponíveis",
			slog.String("account_id", accountID.String()),
			slog.String("campaign_id", campaignID.String()),
			slog.String("error", err.Error()))
		return nil, fmt.Errorf("erro ao contar contatos disponíveis: %w", err)
	}

	// Calcular total de páginas
	totalPages := int(math.Ceil(float64(totalRecords) / float64(perPage)))

	// Aplicar ordenação
	if sort != "" {
		baseQuery += fmt.Sprintf(" ORDER BY %s", sort)
	} else {
		baseQuery += " ORDER BY updated_at DESC"
	}

	r.log.Debug("Query de busca de contatos disponíveis",
		slog.String("query", baseQuery))

	// Aplicar paginação
	offset := (currentPage - 1) * perPage
	baseQuery += fmt.Sprintf(" LIMIT %d OFFSET %d", perPage, offset)

	// Executar query de busca
	rows, err := r.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar contatos disponíveis: %w", err)
	}
	defer rows.Close()

	// Preencher os resultados
	var contacts []models.Contact
	for rows.Next() {
		var contact models.Contact
		var tagsJSON []byte

		if err := rows.Scan(
			&contact.ID, &contact.Name, &contact.Email, &contact.WhatsApp, &contact.Gender,
			&contact.BirthDate, &contact.Bairro, &contact.Cidade, &contact.Estado, &tagsJSON,
			&contact.LastContactAt, &contact.CreatedAt, &contact.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear contatos disponíveis: %w", err)
		}

		// Decodificar JSONB para Tags
		_ = json.Unmarshal(tagsJSON, &contact.Tags)

		contacts = append(contacts, contact)
	}

	// Retornar página de resultados
	return &models.Paginator{
		TotalRecords: totalRecords,
		TotalPages:   totalPages,
		CurrentPage:  currentPage,
		PerPage:      perPage,
		Data:         contacts,
	}, nil
}

// 📌 Buscar contato por WhatsApp ou criar se não existir
func (r *contactRepo) FindOrCreateByWhatsApp(ctx context.Context, accountID uuid.UUID, contact *models.Contact) (*models.Contact, error) {
	query := `
		SELECT id, account_id, name, whatsapp, email, gender, birth_date, bairro, cidade, estado, tags, history, created_at, updated_at
		FROM contacts
		WHERE account_id = $1 AND whatsapp = $2
		LIMIT 1
	`

	var tagsRaw []byte
	err := r.db.QueryRowContext(ctx, query, accountID, contact.WhatsApp).Scan(
		&contact.ID,
		&contact.AccountID,
		&contact.Name,
		&contact.WhatsApp,
		&contact.Email,
		&contact.Gender,
		&contact.BirthDate,
		&contact.Bairro,
		&contact.Cidade,
		&contact.Estado,
		&tagsRaw,
		&contact.History,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		insert := `
			INSERT INTO contacts (
				account_id, name, whatsapp, email, gender, birth_date,
				bairro, cidade, estado, tags, history, created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11, now(), now()
			)
			RETURNING id, created_at, updated_at
		`

		var tagsJSON []byte
		if contact.Tags != nil {
			tagsJSON, _ = json.Marshal(contact.Tags) // se falhar, insere null
		}

		nome := contact.Name
		if nome == "" {
			if contact.WhatsApp != nil && len(*contact.WhatsApp) >= 4 {
				nome = "Contato " + (*contact.WhatsApp)[len(*contact.WhatsApp)-4:]
			} else {
				nome = "Contato"
			}
		}

		err := r.db.QueryRowContext(ctx, insert,
			accountID,
			nome,
			contact.WhatsApp,
			contact.Email,
			contact.Gender,
			contact.BirthDate,
			contact.Bairro,
			contact.Cidade,
			contact.Estado,
			tagsJSON,
			contact.History,
		).Scan(
			&contact.ID,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erro ao criar contato: %w", err)
		}

		contact.AccountID = accountID
		contact.Name = nome
		return contact, nil

	} else if err != nil {
		return nil, fmt.Errorf("erro ao buscar contato: %w", err)
	}

	// Desserializa tags se presentes
	if len(tagsRaw) > 0 {
		var tags models.ContactTags
		if err := json.Unmarshal(tagsRaw, &tags); err == nil {
			contact.Tags = &tags
		}
	}

	return contact, nil
}
