// File: internal/db/postgres/account_otp_repo.go

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type AccountOTPRepoPostgres struct {
	log *slog.Logger
	db  *sql.DB
}

func NewAccountOTPRepository(db *sql.DB) db.AccountOTPRepository {
	return &AccountOTPRepoPostgres{log: logger.GetLogger(), db: db}
}

// FindValidOTP busca um OTP válido para o identificador (e-mail ou WhatsApp)
func (r *AccountOTPRepoPostgres) FindValidOTP(ctx context.Context, identifier string, otp string) (*uuid.UUID, error) {
	var accountID uuid.UUID
	var expiresAt time.Time

	query := `
		SELECT o.account_id, o.expires_at 
		FROM account_otps o
		JOIN accounts a ON o.account_id = a.id
		WHERE o.otp_code = $1 AND (a.email = $2 OR a.whatsapp = $2)
	`
	err := r.db.QueryRow(query, otp, identifier).Scan(&accountID, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("código inválido ou expirado")
		}
		return nil, err
	}

	// Verificar se o OTP expirou
	if time.Now().UTC().After(expiresAt.UTC()) {
		return nil, errors.New("código expirado")
	}

	return &accountID, nil
}

// CleanExpiredOTPs remove registros de OTPs expirados
func (r *AccountOTPRepoPostgres) CleanExpiredOTPs(ctx context.Context) error {
	_, err := r.db.Exec("DELETE FROM account_otps WHERE expires_at < now()")
	return err
}

// FindByEmailOrWhatsApp busca uma conta pelo e-mail ou WhatsApp
func (r *AccountOTPRepoPostgres) FindByEmailOrWhatsApp(ctx context.Context, identifier string) (*models.Account, error) {
	query := "SELECT id, name, email, whatsapp FROM accounts WHERE email = $1 OR whatsapp = $1"
	row := r.db.QueryRow(query, identifier)

	account := &models.Account{}
	err := row.Scan(&account.ID, &account.Name, &account.Email, &account.WhatsApp)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("usuário não encontrado")
		}
		return nil, err
	}
	return account, nil
}

func (r *AccountOTPRepoPostgres) StoreOTP(ctx context.Context, accountID string, otp string) error {
	// 🔥 Garantir que a expiração está no formato UTC
	expiration := time.Now().UTC().Add(10 * time.Minute)

	query := "INSERT INTO account_otps (account_id, otp_code, expires_at) VALUES ($1, $2, $3)"
	_, err := r.db.Exec(query, accountID, otp, expiration)
	return err
}

// GetOTPAttempts retorna o número de tentativas para um OTP específico
func (r *AccountOTPRepoPostgres) GetOTPAttempts(ctx context.Context, identifier string) (int, error) {
	var attempts int
	query := `
        SELECT attempts FROM account_otps
        WHERE account_id = (SELECT id FROM accounts WHERE email = $1 OR whatsapp = $1)
          AND expires_at > NOW()
        ORDER BY created_at DESC
        LIMIT 1
    `
	err := r.db.QueryRowContext(ctx, query, identifier).Scan(&attempts)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil // Se não encontrou, retorna 0 tentativas
		}
		return 0, err
	}
	return attempts, nil
}

// IncrementOTPAttempts adiciona 1 tentativa ao OTP mais recente
func (r *AccountOTPRepoPostgres) IncrementOTPAttempts(ctx context.Context, identifier string) error {
	r.log.Debug("Incrementando tentativas", "identifier", identifier)
	query := `
        UPDATE account_otps
        SET attempts = attempts + 1
        WHERE account_id = (SELECT id FROM accounts WHERE email = $1 OR whatsapp = $1)
    `

	r.log.Debug("Executando query", "query", query)

	_, err := r.db.ExecContext(ctx, query, identifier)
	return err
}

// ResetOTPAttempts reseta as tentativas após uma verificação bem-sucedida
func (r *AccountOTPRepoPostgres) ResetOTPAttempts(ctx context.Context, identifier string) error {
	query := `
        UPDATE account_otps
        SET attempts = 0
        WHERE account_id = (SELECT id FROM accounts WHERE email = $1 OR whatsapp = $1)
    `
	_, err := r.db.ExecContext(ctx, query, identifier)
	return err
}
