// File: /internal/utils/utils.go

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
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

// SafeStringMap evita nil pointer dereference ao acessar mapas de strings opcionais.
func SafeStringMap(m map[string]*string, key string) string {
	if m[key] == nil {
		return "N/A"
	}
	return *m[key]
}

// SendError envia um erro HTTP com o código e mensagem fornecidos.
func SanitizeJSONResponse(rawJSON string) string {
	rawJSON = strings.TrimSpace(rawJSON)               // Remove espaços extras
	rawJSON = strings.TrimPrefix(rawJSON, "```json\n") // Remove cabeçalho errado
	rawJSON = strings.TrimSuffix(rawJSON, "\n```")     // Remove final errado
	return rawJSON
}

// SendError envia um erro HTTP com o código e mensagem fornecidos.
func GetUUIDFromRequestPath(r *http.Request, w http.ResponseWriter, variable string) uuid.UUID {
	idParam := r.PathValue(variable)
	id, err := uuid.Parse(idParam)
	if err != nil {
		SendError(w, http.StatusBadRequest, fmt.Sprintf("%s inválido: %s", variable, idParam))
		return uuid.Nil
	}

	return id
}

// ParseDate parseia uma string de data no formato "yyyy-mm-dd".
func ParseDate(date string) (time.Time, error) {
	return time.Parse("2006-01-02", date)
}
