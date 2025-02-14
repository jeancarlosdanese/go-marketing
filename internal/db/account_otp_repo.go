// File: internal/db/account_otp_repo.go

package db

import (
	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

type AccountOTPRepository interface {
	FindValidOTP(identifier string, otp string) (*uuid.UUID, error)
	CleanExpiredOTPs() error
	FindByEmailOrWhatsApp(identifier string) (*models.Account, error)
	StoreOTP(accountID string, otp string) error
}
