// File: /internal/models/contact_import.go

package models

import (
	"time"

	"github.com/google/uuid"
)

// ContactImport representa um registro de importação de contatos
type ContactImport struct {
	ID        uuid.UUID             `json:"id"`
	AccountID uuid.UUID             `json:"account_id"`
	FileName  string                `json:"file_name"`
	Status    string                `json:"status"`
	Config    *ContactImportConfig  `json:"config,omitempty"`
	Preview   *ContactImportPreview `json:"preview,omitempty"`
	CreatedAt time.Time             `json:"created_at"`
	UpdatedAt time.Time             `json:"updated_at"`
}

// ContactImportConfig define as instruções para como a AI deve agir na importação de contatos
type ContactImportConfig struct {
	AboutData     FieldMapping `json:"about_data"`      // Informações gerais sobre os dados
	Name          FieldMapping `json:"name"`            // Nome do contato
	Email         FieldMapping `json:"email"`           // E-mail do contato
	WhatsApp      FieldMapping `json:"whatsapp"`        // Número de telefone para WhatsApp
	Gender        FieldMapping `json:"gender"`          // Gênero do contato
	BirthDate     FieldMapping `json:"birth_date"`      // Data de nascimento no formato "YYYY-MM-DD"
	Bairro        FieldMapping `json:"bairro"`          // Bairro onde reside
	Cidade        FieldMapping `json:"cidade"`          // Cidade onde reside
	Estado        FieldMapping `json:"estado"`          // Sigla do estado (UF)
	Interesses    FieldMapping `json:"interesses"`      // Como a AI deve categorizar os interesses
	Perfil        FieldMapping `json:"perfil"`          // Como a AI deve definir o perfil
	Eventos       FieldMapping `json:"eventos"`         // Como a AI deve categorizar os eventos
	History       FieldMapping `json:"history"`         // Como a AI deve gerar o histórico
	LastContactAt FieldMapping `json:"last_contact_at"` // Como a AI deve definir a última data de contato
}

// FieldMapping define como um campo do CSV deve ser interpretado pela IA
type FieldMapping struct {
	Source string `json:"source"` // Nome da coluna no CSV que contém essa informação
	Rules  string `json:"rules"`  // Regras para a IA interpretar esse campo
}

// ContactImportPreview define a estrutura do preview do CSV
type ContactImportPreview struct {
	Headers []string   `json:"headers"` // Cabeçalhos do CSV
	Rows    [][]string `json:"rows"`    // Primeiras linhas do CSV
}
