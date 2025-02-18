// File: /internal/server/handlers/import_handler.go

package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"

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
	log  *slog.Logger
	repo db.ContactRepository
}

// NewImportHandler cria uma instância do ImportHandler
func NewImportHandler(repo db.ContactRepository) ImportHandler {
	return &importHandler{
		log:  logger.GetLogger(),
		repo: repo,
	}
}

// 📌 Upload do CSV e processamento direto para o banco de dados
func (h *importHandler) UploadCSVHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount, ok := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)
		if !ok {
			return
		}

		// 📌 Ler o JSON de configuração enviado no campo "config"
		configJSON := r.FormValue("config")
		if configJSON == "" {
			h.log.Error("Configuração de importação ausente")
			utils.SendError(w, http.StatusBadRequest, "É necessário enviar a configuração de importação.")
			return
		}

		var config *dto.ConfigImportContactDTO
		if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
			h.log.Error("Erro ao decodificar configuração de importação", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Configuração de importação inválida.")
			return
		}

		// 📌 Lendo o arquivo CSV do request
		file, _, err := r.FormFile("file")
		if err != nil {
			h.log.Error("Erro ao receber arquivo CSV", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao receber arquivo CSV.")
			return
		}
		defer file.Close()

		// 📌 Criar um arquivo temporário para armazenar o CSV
		tempFilePath := fmt.Sprintf("tmp/uploaded_%s.csv", authAccount.ID)
		outFile, err := os.Create(tempFilePath)
		if err != nil {
			h.log.Error("Erro ao criar arquivo temporário", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao processar CSV.")
			return
		}
		defer outFile.Close()

		// 📌 Processar CSV e salvar no banco usando a configuração definida pelo usuário
		successCount, failedCount, err := service.ProcessCSVAndSaveDB(file, h.repo, authAccount.ID, config)
		if err != nil {
			h.log.Error("Erro ao processar CSV", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao processar CSV.")
			return
		}

		// 📌 Retornar resumo da importação
		response := map[string]interface{}{
			"message":      "Importação concluída!",
			"successCount": successCount,
			"failedCount":  failedCount,
		}

		h.log.Info("CSV processado com sucesso", "account_id", authAccount.ID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
