// File: /internal/server/handlers/account_settings_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type AccountSettingsHandle interface {
	CreateAccountSettingsHandler() http.HandlerFunc
	GetAccountSettingsHandler() http.HandlerFunc
	UpdateAccountSettingsHandler() http.HandlerFunc
	DeleteAccountSettingsHandler() http.HandlerFunc
}

type accountSettingsHandle struct {
	log  *slog.Logger
	repo db.AccountSettingsRepository
}

func NewAccountSettingsHandle(repo db.AccountSettingsRepository) AccountSettingsHandle {
	return &accountSettingsHandle{
		log:  logger.GetLogger(),
		repo: repo,
	}
}

func (h *accountSettingsHandle) CreateAccountSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var settingsDTO dto.AccountSettingsCreateDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&settingsDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// Se não for admin ou (settingsDTO.AccountID == nil), define `AccountID` automaticamente
		if !authAccount.IsAdmin() || settingsDTO.AccountID == nil {
			settingsDTO.AccountID = &authAccount.ID
		}

		// Validar DTO
		if err := settingsDTO.Validate(authAccount.IsAdmin()); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Criar configurações no banco
		settings := &models.AccountSettings{
			AccountID:          *settingsDTO.AccountID,
			OpenAIAPIKey:       settingsDTO.OpenAIAPIKey,
			EvolutionInstance:  settingsDTO.EvolutionInstance,
			AWSAccessKeyID:     settingsDTO.AWSAccessKeyID,
			AWSSecretAccessKey: settingsDTO.AWSSecretAccessKey,
			AWSRegion:          settingsDTO.AWSRegion,
			MailFrom:           settingsDTO.MailFrom,
			MailAdminTo:        settingsDTO.MailAdminTo,
		}

		createdSettings, err := h.repo.Create(r.Context(), settings)
		if err != nil {
			h.log.Error("Erro ao criar configurações da conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao criar configurações")
			return
		}

		h.log.Info("Configurações criadas com sucesso", "account_id", createdSettings.AccountID)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(dto.NewAccountSettingsResponseDTO(createdSettings))
	}
}

// GetAccountSettingsHandler busca as configurações de uma conta.
func (h *accountSettingsHandle) GetAccountSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar `account_id` da URL (se existir)
		accountIDParam := r.PathValue("account_id")
		var accountID uuid.UUID
		var err error

		if accountIDParam == "" {
			// Se nenhum `account_id` for passado, assumimos o do próprio usuário autenticado
			accountID = authAccount.ID
		} else {
			// Verifica se o `account_id` na URL é válido
			accountID, err = uuid.Parse(accountIDParam)
			if err != nil {
				h.log.Warn("ID inválido informado", "account_id", accountIDParam)
				utils.SendError(w, http.StatusBadRequest, "ID inválido")
				return
			}

			// ⚠️ Se não for Admin, não pode buscar configurações de outra conta
			if !authAccount.IsAdmin() && authAccount.ID != accountID {
				h.log.Warn("Usuário tentou acessar configurações de outra conta", "user_id", authAccount.ID, "requested_id", accountID)
				utils.SendError(w, http.StatusForbidden, "Acesso negado")
				return
			}
		}

		// 🔍 Buscar configurações da conta
		settings, err := h.repo.GetByAccountID(r.Context(), accountID)
		if err != nil {
			h.log.Error("Erro ao buscar configurações da conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar configurações")
			return
		}
		if settings == nil {
			utils.SendError(w, http.StatusNotFound, "Nenhuma configuração encontrada")
			return
		}

		h.log.Info("Configurações recuperadas com sucesso", "account_id", accountID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewAccountSettingsResponseDTO(settings))
	}
}

func (h *accountSettingsHandle) UpdateAccountSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var settingsDTO dto.AccountSettingsUpdateDTO

		// Decodifica JSON
		if err := json.NewDecoder(r.Body).Decode(&settingsDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// Verifica se é admin e define o `accountID`
		var accountID uuid.UUID
		if authAccount.IsAdmin() {
			if settingsDTO.AccountID == nil {
				h.log.Warn("Admin deve informar account_id para atualizar configurações")
				utils.SendError(w, http.StatusBadRequest, "account_id é obrigatório para administradores")
				return
			}
			accountID = *settingsDTO.AccountID
		} else {
			if settingsDTO.AccountID != nil {
				h.log.Warn("Usuário comum não pode alterar account_id")
				utils.SendError(w, http.StatusForbidden, "Usuário comum não pode alterar account_id")
				return
			}
			accountID = authAccount.ID
		}

		// 🔍 Buscar configurações da conta
		existingSettings, err := h.repo.GetByAccountID(r.Context(), accountID)
		if err != nil {
			h.log.Error("Erro ao buscar configurações da conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar configurações")
			return
		}
		if existingSettings == nil {
			h.log.Warn("Nenhuma configuração encontrada para a conta")
			utils.SendError(w, http.StatusNotFound, "Nenhuma configuração encontrada para esta conta")
			return
		}

		// Validar DTO
		if err := settingsDTO.Validate(authAccount.IsAdmin()); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// 🔄 Atualizar somente os campos informados na requisição
		if settingsDTO.OpenAIAPIKey != nil {
			existingSettings.OpenAIAPIKey = *settingsDTO.OpenAIAPIKey
		}
		if settingsDTO.EvolutionInstance != nil {
			existingSettings.EvolutionInstance = *settingsDTO.EvolutionInstance
		}
		if settingsDTO.AWSAccessKeyID != nil {
			existingSettings.AWSAccessKeyID = *settingsDTO.AWSAccessKeyID
		}
		if settingsDTO.AWSSecretAccessKey != nil {
			existingSettings.AWSSecretAccessKey = *settingsDTO.AWSSecretAccessKey
		}
		if settingsDTO.AWSRegion != nil {
			existingSettings.AWSRegion = *settingsDTO.AWSRegion
		}
		if settingsDTO.MailFrom != nil {
			existingSettings.MailFrom = *settingsDTO.MailFrom
		}
		if settingsDTO.MailAdminTo != nil {
			existingSettings.MailAdminTo = *settingsDTO.MailAdminTo
		}

		// 🔄 Atualizar configurações no banco
		updatedSettings, err := h.repo.UpdateByAccountID(r.Context(), accountID, existingSettings)
		if err != nil {
			h.log.Error("Erro ao atualizar configurações da conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar configurações")
			return
		}

		h.log.Info("Configurações atualizadas com sucesso", "account_id", updatedSettings.AccountID)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewAccountSettingsResponseDTO(updatedSettings))
	}
}

// DeleteAccountSettingsHandler remove as configurações de uma conta.
func (h *accountSettingsHandle) DeleteAccountSettingsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar conta autenticada
		authAccount := middleware.GetAuthAccountOrFail(r.Context(), w, h.log)

		// 🔍 Capturar `account_id` da URL (se existir)
		accountIDParam := r.PathValue("account_id")
		var accountID uuid.UUID
		var err error

		if accountIDParam == "" {
			// Se nenhum `account_id` for passado, assumimos o do próprio usuário autenticado
			accountID = authAccount.ID
		} else {
			// Verifica se o `account_id` na URL é válido
			accountID, err = uuid.Parse(accountIDParam)
			if err != nil {
				h.log.Warn("ID inválido informado", "account_id", accountIDParam)
				utils.SendError(w, http.StatusBadRequest, "ID inválido")
				return
			}

			// ⚠️ Se não for Admin, não pode buscar configurações de outra conta
			if !authAccount.IsAdmin() && authAccount.ID != accountID {
				h.log.Warn("Usuário tentou acessar configurações de outra conta", "user_id", authAccount.ID, "requested_id", accountID)
				utils.SendError(w, http.StatusForbidden, "Acesso negado")
				return
			}
		}

		// ❌ Deletar configurações
		err = h.repo.DeleteByAccountID(r.Context(), accountID)
		if err != nil {
			h.log.Error("Erro ao deletar configurações da conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao deletar configurações")
			return
		}

		h.log.Info("Configurações deletadas com sucesso", "account_id", authAccount.ID)
		w.WriteHeader(http.StatusNoContent)
	}
}
