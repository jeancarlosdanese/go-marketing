// File: /internal/utils/utils.go

package utils

import (
	"encoding/json"
	"strings"
)

// StrPtr retorna um ponteiro para a string fornecida
func StrPtr(s string) *string {
	return &s
}

// IsValidJSON verifica se uma string é um JSON válido.
func IsValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

// SafeString evita nil pointer dereference ao acessar strings opcionais.
func SafeString(s *string) string {
	if s == nil {
		return "N/A"
	}
	return *s
}

func SanitizeJSONResponse(rawJSON string) string {
	rawJSON = strings.TrimSpace(rawJSON)               // Remove espaços extras
	rawJSON = strings.TrimPrefix(rawJSON, "```json\n") // Remove cabeçalho errado
	rawJSON = strings.TrimSuffix(rawJSON, "\n```")     // Remove final errado
	return rawJSON
}
