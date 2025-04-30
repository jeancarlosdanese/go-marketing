// File: internal/utils/file_utils.go

package utils

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
)

func SaveBytes(data []byte, filename string) error {
	storagePath := os.Getenv("CONTACT_IMPORT_STORAGE_PATH")
	if storagePath == "" {
		return os.ErrNotExist
	}

	fullPath := filepath.Join(storagePath, filename)
	return os.WriteFile(fullPath, data, 0644)
}

// OpenImportContactsFile abre o arquivo salvo da importação
func OpenImportContactsFile(filename string) (multipart.File, error) {
	storagePath := os.Getenv("CONTACT_IMPORT_STORAGE_PATH")
	if storagePath == "" {
		return nil, fmt.Errorf("variável de ambiente CONTACT_IMPORT_STORAGE_PATH não definida")
	}

	fullPath := filepath.Join(storagePath, filename)

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo: %w", err)
	}

	return file, nil
}
