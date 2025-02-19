// File: /internal/utils/utils.go

package utils

import "encoding/json"

// StrPtr retorna um ponteiro para a string fornecida
func StrPtr(s string) *string {
	return &s
}

// IsValidJSON verifica se uma string é um JSON válido.
func IsValidJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
