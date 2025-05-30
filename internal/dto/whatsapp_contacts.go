// internal/models/whatsapp_contacts.go

package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type WhatsappContactDTO struct {
	ID         uuid.UUID `json:"id"`
	ContactID  uuid.UUID `json:"contact_id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	JID        string    `json:"jid"`
	IsBusiness bool      `json:"is_business"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ResolveNumberResponse struct {
	Found           bool                    `json:"found"`
	Input           string                  `json:"input"`
	ResolvedNumber  string                  `json:"resolvedNumber,omitempty"`
	RegisteredJID   string                  `json:"registeredJid,omitempty"`
	DisplayName     string                  `json:"displayName,omitempty"`
	IsBusiness      bool                    `json:"isBusiness"`
	BusinessProfile *models.BusinessProfile `json:"businessProfile,omitempty"`
}

// NewWhatsappContactDTO converte um modelo WhatsappContact em um DTO WhatsappContactDTO
func NewWhatsappContactDTO(whatsappContact models.WhatsappContact) WhatsappContactDTO {
	return WhatsappContactDTO{
		ID:         whatsappContact.ID,
		ContactID:  whatsappContact.ContactID,
		Name:       whatsappContact.Name,
		Phone:      utils.FormatWhatsApp(whatsappContact.Phone),
		JID:        whatsappContact.JID,
		IsBusiness: whatsappContact.IsBusiness,
		UpdatedAt:  whatsappContact.UpdatedAt,
	}
}

// NewWhatsappContactDTOs converte um slice de WhatsappContact em um slice de WhatsappContactDTO
func NewWhatsappContactDTOs(whatsappContacts []models.WhatsappContact) []WhatsappContactDTO {
	var dtos []WhatsappContactDTO
	for _, whatsappContact := range whatsappContacts {
		dtos = append(dtos, NewWhatsappContactDTO(whatsappContact))
	}
	return dtos
}
