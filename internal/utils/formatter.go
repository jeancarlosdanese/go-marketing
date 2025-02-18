// File: /internal/utils/formatter.go

package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// FormatWhatsApp formata um número de telefone para o padrão internacional do WhatsApp
func FormatWhatsApp(number string) string {
	// Remove todos os caracteres não numéricos
	re := regexp.MustCompile(`\D`)
	number = re.ReplaceAllString(number, "")

	// Se for um número brasileiro sem código de país, adiciona +55
	if len(number) == 10 || len(number) == 11 {
		number = "55" + number
	}

	// Verifica se tem código de país
	var countryCode, localNumber string

	if len(number) > 10 { // Número com código de país
		countryCode = number[:2] // Assume que os códigos de país têm 2 dígitos (ajustável)
		localNumber = number[2:]
	} else { // Caso não tenha, assume Brasil
		countryCode = "55"
		localNumber = number
	}

	// Remove prefixos desnecessários
	localNumber = strings.TrimPrefix(localNumber, "0")

	// Aplicando regras de formatação por país
	switch countryCode {
	case "54": // Argentina
		if len(localNumber) >= 10 && localNumber[0] != '9' {
			localNumber = "9" + localNumber[:len(localNumber)-1]
		}
	case "52": // México
		if len(localNumber) >= 10 {
			localNumber = "1" + localNumber
		}
	}

	// Formatação final
	if len(localNumber) == 10 {
		return fmt.Sprintf("+%s %s %s-%s", countryCode, localNumber[:2], localNumber[2:6], localNumber[6:])
	} else if len(localNumber) == 11 {
		return fmt.Sprintf("+%s %s %s-%s", countryCode, localNumber[:2], localNumber[2:7], localNumber[7:])
	} else {
		return fmt.Sprintf("+%s %s", countryCode, localNumber) // Fallback
	}
}

// FormatWhatsAppOnlyNumbers verifica se o número de WhatsApp contém apenas números válidos após limpeza
func FormatWhatsAppOnlyNumbers(whatsapp *string) *string {
	if whatsapp == nil || *whatsapp == "" {
		return nil
	}

	// Remove todos os caracteres não numéricos
	re := regexp.MustCompile(`\D`)
	cleanNumber := re.ReplaceAllString(*whatsapp, "")

	return &cleanNumber
}

// NormalizeEmail normaliza um e-mail para minúsculas e remove espaços extras
func NormalizeEmail(email *string) *string {
	if email == nil || *email == "" {
		return nil
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(*email))

	return &normalizedEmail
}
