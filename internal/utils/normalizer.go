// File: /internal/utils/normalizer.go

package utils

import (
	"strings"
	"time"
	"unicode"

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

// SanitizeWhatsApp remove todos os caracteres não numéricos do telefone
func SanitizeWhatsApp(phone *string) *string {
	if phone == nil {
		return nil
	}
	var result strings.Builder
	for _, r := range *phone {
		if unicode.IsDigit(r) {
			result.WriteRune(r)
		}
	}
	cleaned := result.String()

	// Ajustando formato para 11 dígitos (Brasil)
	if len(cleaned) == 10 { // Caso sem o nono dígito
		cleaned = cleaned[:2] + "9" + cleaned[2:]
	} else if len(cleaned) != 11 {
		return nil // Retorna nil se o número não for válido
	}

	// Adiciona o código do Brasil
	cleaned = "+55 " + cleaned
	return &cleaned
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
