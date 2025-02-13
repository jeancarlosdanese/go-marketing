// File: internal/models/account_settings.go

package models

import "github.com/google/uuid"

// AccountSettings representa as configurações individuais de uma conta
type AccountSettings struct {
	ID                 uuid.UUID `json:"id"`
	AccountID          uuid.UUID `json:"account_id"`
	OpenAIAPIKey       string    `json:"openai_api_key"`        // Máx: 64 chars
	EvolutionAPIURL    string    `json:"evolution_api_url"`     // Máx: 255 chars
	EvolutionAPIKey    string    `json:"evolution_api_key"`     // Máx: 64 chars
	EvolutionInstance  string    `json:"evolution_instance"`    // Máx: 50 chars
	AWSAccessKeyID     string    `json:"aws_access_key_id"`     // Máx: 20 chars
	AWSSecretAccessKey string    `json:"aws_secret_access_key"` // Máx: 40 chars
	AWSRegion          string    `json:"aws_region"`            // Máx: 20 chars
	MailFrom           string    `json:"mail_from"`             // Máx: 150 chars
	MailAdminTo        string    `json:"mail_admin_to"`         // Máx: 150 chars
	XAPIKey            string    `json:"x_api_key"`             // Máx: 64 chars
	LimitDay           int       `json:"limit_day"`
}
