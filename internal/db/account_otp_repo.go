// File: internal/db/account_otp_repo.go

package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type AccountOTPRepository interface {
	FindValidOTP(ctx context.Context, identifier string, otp string) (*uuid.UUID, error)
	CleanExpiredOTPs(ctx context.Context) error
	FindByEmailOrWhatsApp(ctx context.Context, identifier string) (*models.Account, error)
	StoreOTP(ctx context.Context, accountID string, otp string) error
	GetOTPAttempts(ctx context.Context, identifier string) (int, error)
	IncrementOTPAttempts(ctx context.Context, identifier string) error
	ResetOTPAttempts(ctx context.Context, identifier string) error
}
