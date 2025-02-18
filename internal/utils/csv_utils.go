// File: /internal/utils/csv_utils.go

package utils

import (
	"bytes"
	"encoding/csv"
)

// DetectDelimiter tenta identificar automaticamente se o CSV usa `,` ou `;` como delimitador
func DetectDelimiter(data []byte) rune {
	sample := bytes.NewReader(data)
	reader := csv.NewReader(sample)

	// Testa com `;`
	reader.Comma = ';'
	_, err := reader.Read()
	if err == nil {
		return ';'
	}

	// Testa com `,`
	reader.Comma = ','
	_, err = reader.Read()
	if err == nil {
		return ','
	}

	// Retorna `;` como padrão
	return ';'
}
