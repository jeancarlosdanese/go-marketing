// File: internal/dto/account_dto.go

package dto

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/jeancarlosdanese/go-marketing/internal/models"
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
		WhatsApp: formatWhatsApp(account.WhatsApp),
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

// formatWhatsApp formata o número do WhatsApp para exibição no padrão brasileiro e internacional
func formatWhatsApp(number string) string {
	switch len(number) {
	case 10: // Exemplo: "4999999999" (sem código do país)
		return fmt.Sprintf("+55 (%s) %s-%s", number[:2], number[2:6], number[6:])
	case 12: // Exemplo: "554999999999" (com código do país)
		return fmt.Sprintf("+%s (%s) %s-%s", number[:2], number[2:4], number[4:8], number[8:])
	case 13: // Exemplo: "5511999999999" (com código do país)
		return fmt.Sprintf("+%s (%s) %s-%s", number[:2], number[2:4], number[4:9], number[9:])
	default:
		return number // Retorna como está caso o formato não seja reconhecido
	}
}
