// File: internal/models/account_settings.go

package models

import "github.com/google/uuid"

type AccountSettings struct {
	ID                 uuid.UUID `json:"id"`
	AccountID          uuid.UUID `json:"account_id"`
	OpenAIAPIKey       string    `json:"openai_api_key"`
	EvolutionInstance  string    `json:"evolution_instance"`
	AWSAccessKeyID     string    `json:"aws_access_key_id"`
	AWSSecretAccessKey string    `json:"aws_secret_access_key"`
	AWSRegion          string    `json:"aws_region"`
	MailFrom           string    `json:"mail_from"`
	MailAdminTo        string    `json:"mail_admin_to"`
}
