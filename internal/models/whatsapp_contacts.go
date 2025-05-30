// internal/models/whatsapp_contacts.go

package models

import (
	"time"

	"github.com/google/uuid"
)

type WhatsappContact struct {
	ID              uuid.UUID        `json:"id"`
	AccountID       uuid.UUID        `json:"account_id"`
	ContactID       uuid.UUID        `json:"contact_id"`
	Name            string           `json:"name"`
	Phone           string           `json:"phone"`
	JID             string           `json:"jid"`
	IsBusiness      bool             `json:"is_business"`
	BusinessProfile *BusinessProfile `json:"business_profile,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

type BusinessProfile struct {
	Description       *string `json:"description,omitempty"`
	Email             *string `json:"email,omitempty"`
	Address           *string `json:"address,omitempty"`
	Category          *string `json:"category,omitempty"`
	Website           *string `json:"website,omitempty"`
	ProfilePictureURL *string `json:"profile_picture_url,omitempty"`
	CoverPictureURL   *string `json:"cover_picture_url,omitempty"`
	Vertical          *string `json:"vertical,omitempty"`
}
