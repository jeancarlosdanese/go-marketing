// File: /internal/utils/validator.go
package utils

import (
	"errors"
	"regexp"
	"time"
)

// ValidateEmail verifica se o e-mail é válido
func ValidateEmail(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("formato de e-mail inválido")
	}
	return nil
}

// ValidateWhatsApp verifica se o número de WhatsApp contém apenas números válidos após limpeza
func ValidateWhatsApp(whatsapp string) error {
	// Remove todos os caracteres não numéricos
	re := regexp.MustCompile(`\D`)
	cleanNumber := re.ReplaceAllString(whatsapp, "")

	// O número final deve ter entre 10 e 15 dígitos
	if len(cleanNumber) < 10 || len(cleanNumber) > 15 {
		return errors.New("o número de WhatsApp deve ter entre 10 e 15 dígitos válidos")
	}
	return nil
}

// ValidateGender verifica se o gênero informado é válido
func ValidateGender(gender string) error {
	validGenders := map[string]bool{
		"masculino": true,
		"feminino":  true,
		"outro":     true,
	}
	if !validGenders[gender] {
		return errors.New("o gênero deve ser 'masculino', 'feminino' ou 'outro'")
	}
	return nil
}

// ValidateDate verifica se a data está no formato YYYY-MM-DD
func ValidateDate(dateStr string) error {
	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return errors.New("formato de data inválido (use YYYY-MM-DD)")
	}
	return nil
}

// ValidateAWSRegion valida o código da região AWS
func ValidateAWSRegion(region string) error {
	if region == "" {
		return nil
	}

	validRegions := map[string]bool{
		"us-east-1": true, "us-west-1": true, "us-west-2": true, "eu-west-1": true,
		"eu-central-1": true, "ap-southeast-1": true, "ap-southeast-2": true,
		"ap-northeast-1": true, "sa-east-1": true,
	}

	if !validRegions[region] {
		return errors.New("região AWS inválida")
	}
	return nil
}
