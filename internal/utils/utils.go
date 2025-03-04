// File: /internal/utils/utils.go

package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

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

// FileNameNormalize normaliza o nome do arquivo
func FileNameNormalize(originalName string) string {
	// 🔹 Remove a extensão para processar apenas o nome
	nameWithoutExt := strings.TrimSuffix(originalName, ".csv")

	// 🔹 Converte para minúsculas
	normalized := strings.ToLower(nameWithoutExt)

	// 🔹 Remove acentos e normaliza caracteres
	normalized = removeAccents(normalized)

	// 🔹 Substitui espaços por "_"
	normalized = strings.ReplaceAll(normalized, " ", "_")

	// 🔹 Remove caracteres inválidos, mantendo apenas letras, números, `_`, `-`
	reg := regexp.MustCompile(`[^a-z0-9_-]`)
	normalized = reg.ReplaceAllString(normalized, "")

	// 🔹 Remove múltiplos `_` ou `-` consecutivos
	normalized = regexp.MustCompile(`[_-]+`).ReplaceAllString(normalized, "_")

	// 🔹 Garante que o nome não fique muito curto
	if utf8.RuneCountInString(normalized) < 3 {
		normalized = "arquivo"
	}

	// 🔹 Adiciona timestamp e extensão `.csv`
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s_%d.csv", normalized, timestamp)
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
