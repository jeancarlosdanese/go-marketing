// File: /internal/server/handlers/contact_import_handler.go

package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
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
	RemoveImportHandler() http.HandlerFunc
	StartImportHandler() http.HandlerFunc
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

// UploadHandler processa o upload de um arquivo CSV
func (h *importContactHandler) UploadHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, logger.GetLogger())

		// 🔹 Recebe o arquivo
		file, header, err := r.FormFile("file")
		if err != nil {
			h.log.Error("❌ Erro ao receber arquivo", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, "Erro ao receber arquivo.")
			return
		}
		defer file.Close()

		// 🔹 Lê todo o conteúdo do arquivo
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			h.log.Error("❌ Erro ao ler conteúdo do arquivo", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao ler conteúdo do arquivo.")
			return
		}

		// 🔹 Gera nome único e salva no disco
		uniqueFilename := utils.GenerateUniqueFilename(header.Filename)
		if err := utils.SaveBytes(fileBytes, uniqueFilename); err != nil {
			h.log.Error("❌ Erro ao salvar arquivo", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao salvar arquivo.")
			return
		}

		// 🔹 Lê as 5 primeiras linhas do conteúdo para gerar preview
		previewData, err := utils.Read5LinesFromBytes(fileBytes)
		if err != nil || len(previewData) == 0 {
			h.log.Error("❌ Erro ao gerar preview do CSV", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, "Erro ao ler o arquivo CSV. Verifique o formato.")
			return
		}

		preview := &models.ContactImportPreview{
			Headers: previewData[0],
			Rows:    previewData[1:],
		}

		// 🔹 Cria entrada inicial no banco
		importData := &models.ContactImport{
			AccountID: authAccount.ID,
			FileName:  uniqueFilename,
			Status:    "pendente",
			Preview:   preview,
		}
		importRecord, err := h.importContactRepo.Create(r.Context(), importData)
		if err != nil {
			h.log.Error("❌ Erro ao registrar importação no banco", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao registrar importação.")
			return
		}

		// 🔹 Gera configuração automática
		config, err := h.contactImportService.GenerateImportConfig(r.Context(), preview.Headers, preview.Rows)
		if err != nil {
			h.log.Error("❌ Erro ao gerar configuração com IA", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao gerar configuração.")
			return
		}

		_, err = h.importContactRepo.UpdateConfig(r.Context(), authAccount.ID, importRecord.ID, *config)
		if err != nil {
			h.log.Error("❌ Erro ao salvar configuração", slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao salvar configuração.")
			return
		}

		h.log.Info("✅ Upload processado com sucesso", slog.String("import_id", importRecord.ID.String()))
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(importRecord)
	}
}

// GetImportsHandler retorna todas as importações de contatos
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

// GetImportByIDHandler retorna uma importação específica
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

// UpdateImportConfigHandler atualiza a configuração de uma importação
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

// StartImportHandler inicia o processamento de uma importação
func (h *importContactHandler) StartImportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔐 Recupera a conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 📥 Extrai o ID da importação da URL
		importID := utils.GetUUIDFromRequestPath(r, w, "id")

		// 🟡 Atualiza o status para "processando"
		err := h.importContactRepo.UpdateStatus(r.Context(), importID, "processando")
		if err != nil {
			h.log.Error("❌ Erro ao atualizar status para 'processando'", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao iniciar o processamento.")
			return
		}

		// 🚀 Inicia o processamento em background
		go func() {
			ctx := context.Background()

			// 🔍 Recupera os dados da importação
			importData, err := h.importContactRepo.GetByID(ctx, authAccount.ID, importID)
			if err != nil || importData == nil {
				h.log.Error("❌ Erro ao buscar dados da importação", "error", err)
				return
			}

			// 📂 Abre o arquivo CSV
			file, err := utils.OpenImportContactsFile(importData.FileName)
			if err != nil {
				h.log.Error("❌ Erro ao abrir arquivo CSV", "error", err)
				return
			}
			defer file.Close()

			// 🔄 Converte config para DTO
			configDTO := dto.ConvertToConfigImportContactDTO(*importData.Config)

			// 🔧 Processa os dados do CSV
			success, failed, err := h.contactImportService.ProcessCSVAndSaveDB(ctx, file, authAccount.ID, configDTO)
			if err != nil {
				h.log.Error("❌ Erro ao processar CSV", "error", err)
				return
			}

			h.log.Info("✅ Processamento concluído",
				slog.Int("success", success),
				slog.Int("failed", failed),
				slog.String("import_id", importID.String()),
			)

			// ✅ Atualiza o status para "concluida"
			err = h.importContactRepo.UpdateStatus(context.Background(), importID, "concluida")
			if err != nil {
				h.log.Error("❌ Erro ao atualizar status para 'concluida'", slog.String("error", err.Error()))
			}

		}()

		// 🔁 Retorna resposta imediata
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "processando"})
	}
}

// RemoveImportHandler remove uma importação de contatos
func (h *importContactHandler) RemoveImportHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Buscar ID da campanha
		importID := utils.GetUUIDFromRequestPath(r, w, "id")

		// Remover a importação
		err := h.importContactRepo.Remove(r.Context(), authAccount.ID, importID)
		if err != nil {
			utils.SendError(w, http.StatusInternalServerError, "Erro ao remover importação.")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
