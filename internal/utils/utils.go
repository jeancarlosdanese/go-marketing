// File: /internal/utils/utils.go

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

// ExtractPaginationParams extrai os parâmetros de paginação da query string
func ExtractPaginationParams(r *http.Request) (int, int, string) {
	// Padrões caso não sejam passados na URL
	defaultPage := 1
	defaultPerPage := 10
	defaultSort := "updated_at DESC"

	// Extrair `page` da query string (com fallback para o valor padrão)
	pageParam := r.URL.Query().Get("page")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		page = defaultPage
	}

	// Extrair `per_page` da query string (com fallback para o valor padrão)
	perPageParam := r.URL.Query().Get("per_page")
	perPage, err := strconv.Atoi(perPageParam)
	if err != nil || perPage < 1 {
		perPage = defaultPerPage
	}

	// Extrair `sort` da query string (com fallback para o valor padrão)
	sort := r.URL.Query().Get("sort")
	if sort == "" {
		sort = defaultSort
	}

	return page, perPage, sort
}
