// File: /internal/models/paginator.go

package models

// Paginator representa uma página paginada de resultados
type Paginator struct {
	TotalRecords int         `json:"total_records"` // Total de registros na consulta
	TotalPages   int         `json:"total_pages"`   // Total de páginas disponíveis
	CurrentPage  int         `json:"current_page"`  // Página atual
	PerPage      int         `json:"per_page"`      // Registros por página
	Data         interface{} `json:"data"`          // Dados da página atual
}
