// File: internal/utils/import_contacts_manager.go

package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// GenerateUniqueFilename gera um nome de arquivo único baseado no nome original
func GenerateUniqueFilename(originalName string) string {
	return fileNameNormalize(originalName)
}

// GetImportContactsFilePath retorna o caminho do arquivo baseado no TEMPLATE_STORAGE_PATH
func GetImportContactsFilePath(uniqueFilename string) string {
	basePath := os.Getenv("CONTACT_IMPORT_STORAGE_PATH")
	if basePath == "" {
		basePath = "./uploads/contacts" // Fallback
	}

	// Retorna o caminho completo
	return filepath.Join(basePath, uniqueFilename)
}

// SaveImportContactsFile salva o arquivo de importação de contatos no disco
func SaveImportContactsFile(file multipart.File, filename string) error {
	path := GetImportContactsFilePath(filename)

	// Criar diretórios se não existirem
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	// Salvar no disco
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, fileBytes, 0644); err != nil {
		return err
	}

	return nil
}

// 🔹 Lê algumas linhas do CSV (usando bufio para evitar EOF inesperado)
func Read5LinesFromBytes(data []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	var preview [][]string

	for i := 0; i < 5; i++ {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		preview = append(preview, row)
	}
	return preview, nil
}

// DeleteImportContactsFile deleta o arquivo de importação de contatos do disco
func DeleteImportContactsFile(filename string) error {
	path := GetImportContactsFilePath(filename)
	return os.Remove(path)
}

// fileNameNormalize normaliza o nome do arquivo
func fileNameNormalize(originalName string) string {
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
