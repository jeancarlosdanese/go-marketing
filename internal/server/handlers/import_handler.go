// File: /internal/server/handlers/import_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/service"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

// ImportHandler define a interface do handler de importação
type ImportHandler interface {
	UploadCSVHandler() http.HandlerFunc
}

type importHandler struct {
	log                  *slog.Logger
	contatctRepo         db.ContactRepository
	contactImportService service.ContactImportService
}

// NewImportHandler cria uma instância do ImportHandler
func NewImportHandler(contatctRepo db.ContactRepository, openAIService service.OpenAIService) ImportHandler {
	return &importHandler{
		log:                  logger.GetLogger(),
		contatctRepo:         contatctRepo,
		contactImportService: service.NewContactImportService(contatctRepo, openAIService),
	}
}

// 📌 Upload do CSV e processamento direto para o banco de dados
func (h *importHandler) UploadCSVHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 📌 Ler o JSON de configuração enviado no campo "config"
		configJSON := r.FormValue("config")
		if configJSON == "" {
			h.log.Warn("Configuração de importação ausente",
				slog.String("account_id", authAccount.ID.String()))
			utils.SendError(w, http.StatusBadRequest, "É necessário enviar a configuração de importação.")
			return
		}

		var config dto.ConfigImportContactDTO
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			h.log.Warn("Erro ao decodificar configuração de importação",
				slog.String("account_id", authAccount.ID.String()),
				slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, "Configuração de importação inválida.")
			return
		}

		// 📌 Lendo o arquivo CSV do request
		file, _, err := r.FormFile("file")
		if err != nil {
			h.log.Warn("Erro ao receber arquivo CSV",
				slog.String("account_id", authAccount.ID.String()),
				slog.String("error", err.Error()))
			utils.SendError(w, http.StatusBadRequest, "Erro ao receber arquivo CSV.")
			return
		}
		defer file.Close()

		// 📌 Processar CSV e salvar no banco
		successCount, failedCount, err := h.contactImportService.ProcessCSVAndSaveDB(r.Context(), file, authAccount.ID, &config)
		if err != nil {
			h.log.Error("Erro ao processar CSV",
				slog.String("account_id", authAccount.ID.String()),
				slog.String("error", err.Error()))
			utils.SendError(w, http.StatusInternalServerError, "Erro ao processar CSV.")
			return
		}

		// 📌 Resumo da importação
		response := map[string]interface{}{
			"message":      "Importação concluída!",
			"successCount": successCount,
			"failedCount":  failedCount,
		}

		h.log.Info("CSV processado com sucesso",
			slog.String("account_id", authAccount.ID.String()),
			slog.Int("successCount", successCount),
			slog.Int("failedCount", failedCount))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
