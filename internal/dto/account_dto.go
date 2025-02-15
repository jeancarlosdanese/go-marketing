// File: internal/dto/account_dto.go

package dto

import (
	"errors"
	"regexp"

	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// AccountCreateDTO define os campos necessários para criar uma conta
type AccountCreateDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	WhatsApp string `json:"whatsapp"`
}

// AccountUpdateDTO define os campos permitidos para atualizar uma conta
type AccountUpdateDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	WhatsApp string `json:"whatsapp"`
}

// AccountResponseDTO define a estrutura de resposta para a conta
type AccountResponseDTO struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	WhatsApp string `json:"whatsapp"`
}

// Novo construtor para AccountResponseDTO
func NewAccountResponseDTO(account *models.Account) AccountResponseDTO {
	return AccountResponseDTO{
		ID:       account.ID.String(),
		Name:     account.Name,
		Email:    account.Email,
		WhatsApp: utils.FormatWhatsApp(account.WhatsApp),
	}
}

// Validate valida os dados antes de criar uma conta
func (a *AccountCreateDTO) Validate() error {
	// Validar nome
	if len(a.Name) < 3 || len(a.Name) > 100 {
		return errors.New("o nome deve ter entre 3 e 100 caracteres")
	}

	// Validar e-mail
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(a.Email) {
		return errors.New("e-mail inválido")
	}

	// Validar WhatsApp (somente números)
	re := regexp.MustCompile(`\D`) // Remove tudo que não for número
	a.WhatsApp = re.ReplaceAllString(a.WhatsApp, "")

	if len(a.WhatsApp) < 10 || len(a.WhatsApp) > 15 {
		return errors.New("o WhatsApp deve ter entre 10 e 15 dígitos")
	}

	return nil
}

// Validate valida os dados antes de atualizar uma conta
func (a *AccountUpdateDTO) Validate() error {
	// Validar nome se for enviado
	if a.Name != "" && (len(a.Name) < 3 || len(a.Name) > 100) {
		return errors.New("o nome deve ter entre 3 e 100 caracteres")
	}

	// Validar e-mail se for enviado
	if a.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(a.Email) {
			return errors.New("e-mail inválido")
		}
	}

	// Validar WhatsApp se for enviado (somente números)
	if a.WhatsApp != "" {
		re := regexp.MustCompile(`\D`) // Remove tudo que não for número
		a.WhatsApp = re.ReplaceAllString(a.WhatsApp, "")

		if len(a.WhatsApp) < 10 || len(a.WhatsApp) > 15 {
			return errors.New("o WhatsApp deve ter entre 10 e 15 dígitos")
		}
	}

	return nil
}
