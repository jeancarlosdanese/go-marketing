// File: /internal/server/handlers/contact_import_handler.go

package handlers

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// ImportContactHandler define a interface do handler de importação
type ImportContactHandler interface {
	UploadHandler() http.HandlerFunc
	GetImportsHandler() http.HandlerFunc
	GetImportByIDHandler() http.HandlerFunc
	UpdateImportConfigHandler() http.HandlerFunc
}

type importContactHandler struct {
	log                  *slog.Logger
	importContactRepo    db.ContactImportRepository
	contactImportService service.ContactImportService
}

// NewImportHandler cria uma instância do ImportHandler
func NewImportContactHandler(importContactRepo db.ContactImportRepository, contactImportService service.ContactImportService) ImportContactHandler {
	return &importContactHandler{
		log:                  logger.GetLogger(),
		importContactRepo:    importContactRepo,
		contactImportService: contactImportService,
	}
}

// UploadHandler gerencia o upload de arquivos CSV para importação de contatos
func (h *importContactHandler) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, logger.GetLogger())

		// 🔹 Caminho de armazenamento
		storagePath := os.Getenv("CONTACT_IMPORT_STORAGE_PATH")
		if storagePath == "" {
			h.log.Error("❌ Variável de ambiente CONTACT_IMPORT_STORAGE_PATH não definida")
			utils.SendError(w, http.StatusInternalServerError, "Erro interno: caminho de armazenamento não configurado.")
			return
		}

		// 🔹 Recebe o arquivo CSV
		file, header, err := r.FormFile("file")
		if err != nil {
			h.log.Error("❌ Erro ao receber arquivo", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, "Erro ao receber arquivo.")
			return
		}
		defer file.Close()

		h.log.Debug("📁 Arquivo recebido", slog.String("filename", header.Filename), slog.Int64("size", header.Size))

		// 🔹 Gera um nome único para o arquivo
		uniqueFilename := utils.FileNameNormalize(header.Filename)
		savePath := filepath.Join(storagePath, uniqueFilename)

		// 🔹 Salva o arquivo localmente
		outFile, err := os.Create(savePath)
		if err != nil {
			h.log.Error("❌ Erro ao salvar arquivo", slog.String("path", savePath), slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao salvar arquivo.")
			return
		}
		defer outFile.Close()

		if _, err := outFile.ReadFrom(file); err != nil {
			h.log.Error("❌ Erro ao escrever no arquivo", slog.String("path", savePath), slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao escrever arquivo.")
			return
		}

		// 🔹 Reseta a posição do arquivo antes da leitura
		outFile.Seek(0, 0)

		// 🔹 Lê algumas linhas do CSV (usando bufio para evitar EOF inesperado)
		reader := csv.NewReader(bufio.NewReader(outFile))
		previewData := make([][]string, 0)

		for i := 0; i < 5; i++ {
			line, err := reader.Read()
			if err != nil {
				h.log.Error("❌ Erro ao ler linha do CSV", slog.String("error", err.Error()))
				break
			}
			previewData = append(previewData, line)
		}

		// Criando a estrutura do preview corretamente antes de salvar
		preview := &models.ContactImportPreview{
			Headers: previewData[0],  // Primeira linha do CSV contém os cabeçalhos
			Rows:    previewData[1:], // Restante das linhas são os dados
		}

		// Criar o objeto da importação
		importData := &models.ContactImport{
			AccountID: authAccount.ID,
			FileName:  uniqueFilename,
			Status:    "pendente",
			Preview:   preview,
		}

		// 🔹 Salvar no banco
		importRecord, err := h.importContactRepo.Create(r.Context(), importData)
		if err != nil {
			h.log.Error("❌ Erro ao registrar importação no banco", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao registrar importação.")
			return
		}

		// 🔹 Gerar a configuração inicial **sincronamente** (IA roda antes da resposta HTTP)
		config, err := h.contactImportService.GenerateImportConfig(r.Context(), preview.Headers, preview.Rows)
		if err != nil {
			h.log.Error("❌ Erro ao gerar configuração com IA", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao gerar configuração.")
			return
		}

		// 🔹 Atualizar a importação no banco com a configuração gerada
		_, err = h.importContactRepo.UpdateConfig(r.Context(), authAccount.ID, importRecord.ID, *config)
		if err != nil {
			h.log.Error("❌ Erro ao salvar configuração gerada no banco", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao salvar configuração.")
			return
		}

		h.log.Info("✅ Configuração inicial gerada e salva com sucesso", slog.String("import_id", importRecord.ID.String()))

		jsonData, _ := json.Marshal(importRecord)
		h.log.Debug("📝 Importação registrada no banco", "json", jsonData)

		h.log.Info("✅ Upload processado com sucesso", slog.String("import_id", importRecord.ID.String()), slog.String("filename", uniqueFilename))

		// 🔹 Retorna a importação já com a configuração gerada
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(importRecord)
	}
}

func (h *importContactHandler) GetImportsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// Buscar importações para a conta do usuário
		imports, err := h.importContactRepo.GetAllByAccountID(r.Context(), authAccount.ID)
		if err != nil {
			h.log.Error("❌ Erro ao buscar importações", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar importações.")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(imports)
	}
}

func (h *importContactHandler) GetImportByIDHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Buscar ID da campanha
		importID := utils.GetUUIDFromRequestPath(r, w, "id")

		h.log.Debug("🔍 Buscando importação", "import_id", importID)

		// Buscar importação pelo ID
		contactImport, err := h.importContactRepo.GetByID(r.Context(), authAccount.ID, importID)
		if err != nil {
			h.log.Error("❌ Erro ao buscar importação", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar importação.")
			return
		}

		if contactImport == nil {
			utils.SendError(w, http.StatusNotFound, "Importação não encontrada.")
			return
		}

		h.log.Debug("🔍 Importação encontrada", "import", contactImport)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(contactImport)
	}
}

func (h *importContactHandler) UpdateImportConfigHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Buscar ID da campanha
		importID := utils.GetUUIDFromRequestPath(r, w, "id")

		// Decodifica o JSON do request
		var config models.ContactImportConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			utils.SendError(w, http.StatusBadRequest, "Configuração inválida.")
			return
		}

		// Atualiza a configuração no banco de dados
		contactImport, err := h.importContactRepo.UpdateConfig(r.Context(), authAccount.ID, importID, config)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar configuração.")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(contactImport)
	}
}
