// File: /internal/dto/account_settings_dto.go

package dto

import (
	"errors"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// AccountSettingsCreateDTO define os dados necessários para criar configurações de uma conta
type AccountSettingsCreateDTO struct {
	AccountID          *uuid.UUID `json:"account_id,omitempty"` // Admin pode definir
	OpenAIAPIKey       string     `json:"openai_api_key,omitempty"`
	EvolutionInstance  string     `json:"evolution_instance,omitempty"`
	AWSAccessKeyID     string     `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string     `json:"aws_secret_access_key,omitempty"`
	AWSRegion          string     `json:"aws_region,omitempty"`
	MailFrom           string     `json:"mail_from,omitempty"`
	MailAdminTo        string     `json:"mail_admin_to,omitempty"`
}

// AccountSettingsUpdateDTO define os dados permitidos para atualização
type AccountSettingsUpdateDTO struct {
	AccountID          *uuid.UUID `json:"account_id,omitempty"` // Obrigatório para admin
	OpenAIAPIKey       *string    `json:"openai_api_key,omitempty"`
	EvolutionInstance  *string    `json:"evolution_instance,omitempty"`
	AWSAccessKeyID     *string    `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey *string    `json:"aws_secret_access_key,omitempty"`
	AWSRegion          *string    `json:"aws_region,omitempty"`
	MailFrom           *string    `json:"mail_from,omitempty"`
	MailAdminTo        *string    `json:"mail_admin_to,omitempty"`
}

// AccountSettingsResponseDTO estrutura de resposta para configurações de conta
type AccountSettingsResponseDTO struct {
	ID                 string `json:"id"`
	AccountID          string `json:"account_id"`
	OpenAIAPIKey       string `json:"openai_api_key,omitempty"`
	EvolutionInstance  string `json:"evolution_instance,omitempty"`
	AWSAccessKeyID     string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"`
	AWSRegion          string `json:"aws_region,omitempty"`
	MailFrom           string `json:"mail_from,omitempty"`
	MailAdminTo        string `json:"mail_admin_to,omitempty"`
}

// NewAccountSettingsResponseDTO cria um DTO de resposta formatado
func NewAccountSettingsResponseDTO(settings *models.AccountSettings) AccountSettingsResponseDTO {
	return AccountSettingsResponseDTO{
		ID:                 settings.ID.String(),
		AccountID:          settings.AccountID.String(),
		OpenAIAPIKey:       settings.OpenAIAPIKey,
		EvolutionInstance:  settings.EvolutionInstance,
		AWSAccessKeyID:     settings.AWSAccessKeyID,
		AWSSecretAccessKey: settings.AWSSecretAccessKey,
		AWSRegion:          settings.AWSRegion,
		MailFrom:           settings.MailFrom,
		MailAdminTo:        settings.MailAdminTo,
	}
}

// 🔍 Validações

// Validate valida os dados do AccountSettingsCreateDTO
func (a *AccountSettingsCreateDTO) Validate(isAdmin bool) error {
	// Admin deve informar `account_id`, usuários comuns não devem
	if isAdmin {
		if a.AccountID == nil {
			return errors.New("account_id é obrigatório para administradores")
		}
	} else {
		if a.AccountID != nil {
			return errors.New("usuário comum não pode definir account_id")
		}
	}

	// Validações específicas podem ser adicionadas aqui, ex:
	if a.MailFrom == "" && a.MailAdminTo != "" {
		return errors.New("se mail_admin_to for definido, mail_from também deve ser")
	}

	if err := utils.ValidateAWSRegion(a.AWSRegion); err != nil {
		return err
	}
	if err := utils.ValidateEmail(a.MailFrom); err != nil {
		return errors.New("MailFrom inválido: " + err.Error())
	}
	if err := utils.ValidateEmail(a.MailAdminTo); err != nil {
		return errors.New("MailAdminTo inválido: " + err.Error())
	}
	return nil
}

// Validate valida os dados do AccountSettingsUpdateDTO
func (a *AccountSettingsUpdateDTO) Validate(isAdmin bool) error {
	// Se for admin, account_id é obrigatório
	if isAdmin {
		if a.AccountID == nil {
			return errors.New("account_id é obrigatório para administradores")
		}
	} else {
		if a.AccountID != nil {
			return errors.New("usuário comum não pode definir account_id")
		}
	}

	// Validações adicionais
	if a.AWSRegion != nil && *a.AWSRegion != "" {
		if err := utils.ValidateAWSRegion(*a.AWSRegion); err != nil {
			return err
		}
	}
	if a.MailFrom != nil && *a.MailFrom != "" {
		if err := utils.ValidateEmail(*a.MailFrom); err != nil {
			return errors.New("MailFrom inválido: " + err.Error())
		}
	}
	if a.MailAdminTo != nil && *a.MailAdminTo != "" {
		if err := utils.ValidateEmail(*a.MailAdminTo); err != nil {
			return errors.New("MailAdminTo inválido: " + err.Error())
		}
	}
	return nil
}
