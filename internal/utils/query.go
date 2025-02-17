// File: /internal/utils/query.go

package utils

import (
	"net/url"
)

// ExtractQueryFilters extrai os filtros permitidos da query string e retorna um mapa com os valores.
func ExtractQueryFilters(query url.Values, allowedFilters []string) map[string]string {
	filters := make(map[string]string)

	for _, filter := range allowedFilters {
		if value := query.Get(filter); value != "" {
			filters[filter] = value
		}
	}

	return filters
}
