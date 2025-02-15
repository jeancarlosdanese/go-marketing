// File: /internal/utils/template_manager.go

package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// GetTemplatePath retorna o caminho do arquivo baseado no TEMPLATE_STORAGE_PATH
func GetTemplatePath(templateID, templateType string) string {
	basePath := os.Getenv("TEMPLATE_STORAGE_PATH")
	if basePath == "" {
		basePath = "./uploads/templates" // Fallback
	}

	// Define a extensão do arquivo
	ext := "html"
	if templateType == "whatsapp" {
		ext = "md"
	}

	// Retorna o caminho completo
	return filepath.Join(basePath, templateType, fmt.Sprintf("%s.%s", templateID, ext))
}

// SaveTemplate salva o template localmente e faz backup no S3
func SaveTemplate(templateID, templateType string, content []byte) error {
	path := GetTemplatePath(templateID, templateType)

	// Criar diretórios se não existirem
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	// Salvar no disco
	if err := os.WriteFile(path, content, 0644); err != nil {
		return err
	}

	// Backup no S3
	// TODO: return UploadToS3(path) // Implementar

	return nil
}

// LoadTemplate carrega o template do disco ou do S3
func LoadTemplate(templateID, templateType string) ([]byte, error) {
	path := GetTemplatePath(templateID, templateType)

	// Primeiro tenta localmente
	content, err := os.ReadFile(path)
	if err == nil {
		return content, nil
	}

	// Se falhar, tenta recuperar do S3
	err = DownloadFromS3(path)
	if err != nil {
		return nil, errors.New("template não encontrado")
	}

	// Tenta carregar novamente
	return os.ReadFile(path)
}

// DeleteTemplate remove um template do disco e do S3
func DeleteTemplate(templateID, templateType string) error {
	path := GetTemplatePath(templateID, templateType)

	// Remove do disco
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}

	// Remove do S3
	return DeleteFromS3(path)
}

// DownloadFromS3 is a mock implementation of the function to download a file from S3
func DownloadFromS3(path string) error {
	// TODO: Implement the actual S3 download logic
	return nil
}

// UploadToS3 is a mock implementation of the function to upload a file to S3
func UploadToS3(path string) error {
	// TODO: Implement the actual S3 upload logic
	return nil
}

// DeleteFromS3 is a mock implementation of the function to delete a file from S3
func DeleteFromS3(path string) error {
	// TODO: Implement the actual S3 delete logic
	return nil
}
