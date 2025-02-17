// File: /internal/utils/errors.go

package utils

import "strings"

// IsUniqueConstraintError verifica se um erro do banco de dados é uma violação de chave única (UNIQUE constraint).
func IsUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	// Verificar se o erro é uma violação de chave única
	if strings.Contains(err.Error(), "unique") {
		return true
	}

	return false
}
