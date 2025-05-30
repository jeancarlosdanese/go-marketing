// File: /internal/utils/utils.go

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

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

// SanitizeJSONResponse remove espaços extras e formata o JSON.
func SanitizeJSONResponse(rawJSON string) string {
	rawJSON = strings.TrimSpace(rawJSON)               // Remove espaços extras
	rawJSON = strings.TrimPrefix(rawJSON, "```json\n") // Remove cabeçalho errado
	rawJSON = strings.TrimSuffix(rawJSON, "\n```")     // Remove final errado
	return rawJSON
}

// GetUUIDFromRequestPath extrai um UUID de uma variável na URL do request.
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

// removeAccents remove acentos mantendo as letras originais
func removeAccents(input string) string {
	accents := map[rune]rune{
		'á': 'a', 'ã': 'a', 'â': 'a', 'à': 'a', 'ä': 'a',
		'é': 'e', 'ê': 'e', 'è': 'e', 'ë': 'e',
		'í': 'i', 'î': 'i', 'ì': 'i', 'ï': 'i',
		'ó': 'o', 'õ': 'o', 'ô': 'o', 'ò': 'o', 'ö': 'o',
		'ú': 'u', 'û': 'u', 'ù': 'u', 'ü': 'u',
		'ç': 'c',
		'ñ': 'n',
	}

	var output strings.Builder
	for _, r := range input {
		if newR, found := accents[r]; found {
			output.WriteRune(newR)
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '-' || r == '_' {
			output.WriteRune(r)
		}
	}
	return output.String()
}
