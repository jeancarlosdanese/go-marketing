// File: /internal/models/account.go

package models

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// Account representa a entidade de conta no banco de dados
type Account struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	WhatsApp string    `json:"whatsapp"`
}

// Validate verifica se os dados da conta são válidos antes de salvar ou atualizar
func (a *Account) Validate(isUpdate bool) error {
	// Validar nome
	if len(a.Name) == 0 || len(a.Name) > 100 {
		return errors.New("o nome deve ter entre 1 e 100 caracteres")
	}

	// Validar e-mail
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(a.Email) {
		return errors.New("e-mail inválido")
	}

	// 🔥 Limpar o WhatsApp (somente números)
	re := regexp.MustCompile(`\D`) // Remove tudo que não for número
	a.WhatsApp = re.ReplaceAllString(a.WhatsApp, "")

	// Validar WhatsApp
	if len(a.WhatsApp) < 10 || len(a.WhatsApp) > 15 {
		return errors.New("o WhatsApp deve ter entre 10 e 15 dígitos")
	}

	// 🔥 Impedir alteração de e-mail e WhatsApp em updates
	if isUpdate {
		if a.Email != "" || a.WhatsApp != "" {
			return errors.New("e-mail e WhatsApp não podem ser alterados")
		}
	}

	return nil
}

// FormatWhatsApp retorna o número formatado (somente na saída)
func (a *Account) FormatWhatsApp() string {
	if len(a.WhatsApp) == 11 {
		return fmt.Sprintf("+%s (%s) %s-%s", a.WhatsApp[:2], a.WhatsApp[2:4], a.WhatsApp[4:8], a.WhatsApp[8:])
	}
	return a.WhatsApp
}

// IsAdmin verifica se a conta é admin baseada no ID fixo
func (a *Account) IsAdmin() bool {
	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	return a.ID == adminID
}
