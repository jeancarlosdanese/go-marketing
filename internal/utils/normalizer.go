// File: /internal/utils/normalizer.go

package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/nyaruka/phonenumbers"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// NilIfEmpty retorna nil se a string for vazia ou apenas espaços
func NilIfEmpty(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	return &trimmed
}

// Capitalize converte uma string para capitalização correta
func Capitalize(s *string) *string {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil
	}
	caser := cases.Title(language.BrazilianPortuguese)
	capitalized := caser.String(strings.ToLower(strings.TrimSpace(*s)))
	return &capitalized
}

// SanitizeEmail transforma o e-mail em minúsculas e remove espaços extras
func SanitizeEmail(email *string) *string {
	if email == nil {
		return nil
	}
	cleaned := strings.TrimSpace(strings.ToLower(*email))
	if cleaned == "" {
		return nil
	}
	return &cleaned
}

// GetWhatsAppNumber extrai o número de telefone do JID do WhatsApp
func GetWhatsAppOnlyNumber(jid string) string {
	jid = strings.TrimSpace(jid)
	jid = strings.TrimPrefix(jid, "+")
	jid = strings.ReplaceAll(jid, " ", "")
	jid = strings.TrimSuffix(jid, "@s.whatsapp.net")
	jid = OnlyDigits(jid)

	return jid // Retorna como está se não for o formato esperado
}

func NormalizeWhatsAppNumber(jid string) string {
	jid = GetWhatsAppOnlyNumber(jid)

	// Ex: 554999669869 → DDI: 55, DDD: 49, número: 99669869
	if len(jid) == 12 { // Ex: 55 49 99669869 (sem o 9)
		ddi := jid[:2]
		ddd := jid[2:4]
		num := jid[4:]

		// Insere o nono dígito (9) após o DDD
		num = "9" + num

		return ddi + ddd + num // 5549999669869
	}

	return jid // Retorna como está se não for o formato esperado
}

// ParsePhone valida e formata um número de telefone brasileiro
func ParsePhone(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	var num *phonenumbers.PhoneNumber
	var err error

	switch {
	case strings.HasPrefix(raw, "+"):
		num, err = phonenumbers.Parse(raw, "")
	case strings.HasPrefix(raw, "55"):
		num, err = phonenumbers.Parse("+"+raw, "")
	default:
		num, err = phonenumbers.Parse(raw, "BR")
	}

	if err != nil {
		return "", fmt.Errorf("parse error: %w", err)
	}

	if !phonenumbers.IsPossibleNumber(num) {
		return "", fmt.Errorf("número impossível")
	}

	// Sucesso: retorna 5549999669869
	e164 := phonenumbers.Format(num, phonenumbers.E164)
	return strings.TrimPrefix(e164, "+"), nil
}

// NormalizeGender normaliza e valida o gênero (masculino, feminino, outro)
func NormalizeGender(gender *string) *string {
	if gender == nil {
		return nil
	}
	g := strings.ToLower(strings.TrimSpace(*gender))
	switch g {
	case "masculino", "feminino", "outro":
		return &g
	case "m":
		return StrPtr("masculino")
	case "f":
		return StrPtr("feminino")
	default:
		return nil
	}
}

// NormalizeBirthDate normaliza e valida a data de nascimento
func NormalizeBirthDate(date *string) *string {
	if date == nil {
		return nil
	}
	parsed, err := time.Parse("2006-01-02", *date)
	if err != nil || parsed.Year() < 1900 || parsed.Year() > time.Now().Year() {
		return nil
	}
	return date
}

// NormalizeEmail normaliza um e-mail para minúsculas e remove espaços extras
func NormalizeEmail(email *string) *string {
	if email == nil || *email == "" {
		return nil
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(*email))

	return &normalizedEmail
}
