// File: internal/server/handlers/account_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
	"github.com/jeancarlosdanese/go-marketing/internal/utils"
)

type AccountHandle interface {
	CreateAccountHandler() http.HandlerFunc
	GetAllAccountsHandler() http.HandlerFunc
	GetAccountHandler() http.HandlerFunc
	UpdateAccountHandler() http.HandlerFunc
	DeleteAccountHandler() http.HandlerFunc
}

type accountHandle struct {
	log  *slog.Logger
	repo db.AccountRepository
}

func NewAccountHandle(repo db.AccountRepository) AccountHandle {
	return &accountHandle{
		log:  logger.GetLogger(),
		repo: repo,
	}
}

// CreateAccountHandler cria uma nova conta
func (h *accountHandle) CreateAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var accountDTO dto.AccountCreateDTO

		h.log.Debug("Recebendo requisição para criar conta", "method", r.Method, "route", r.URL.Path)

		if err := json.NewDecoder(r.Body).Decode(&accountDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		h.log.Debug("Payload recebido", "name", accountDTO.Name, "email", accountDTO.Email)

		if err := accountDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		account := &models.Account{
			Name:     accountDTO.Name,
			Email:    accountDTO.Email,
			WhatsApp: accountDTO.WhatsApp,
		}

		createdAccount, err := h.repo.Create(account)
		if err != nil {
			h.log.Error("Erro ao criar conta", "error", err)
			if strings.Contains(err.Error(), "duplicate key value") {
				utils.SendError(w, http.StatusConflict, "E-mail ou WhatsApp já cadastrado")
			} else {
				utils.SendError(w, http.StatusInternalServerError, "Erro ao criar conta")
			}
			return
		}

		response := dto.NewAccountResponseDTO(createdAccount)

		h.log.Info("Conta criada com sucesso", "account_id", createdAccount.ID.String(), "email", createdAccount.Email)
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

// GetAllAccountsHandler retorna todas as contas cadastradas
func (h *accountHandle) GetAllAccountsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			utils.SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		if !authAccount.IsAdmin() {
			h.log.Warn("Apenas administradores podem buscar todas as contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem buscar todas as contas")
			return
		}

		accounts, err := h.repo.GetAll()
		if err != nil {
			h.log.Error("Erro ao buscar contas", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao buscar contas")
			return
		}

		h.log.Debug("Contas recuperadas", "count", len(accounts))

		var response []dto.AccountResponseDTO
		for _, acc := range accounts {
			response = append(response, dto.NewAccountResponseDTO(acc))
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// GetAccountHandler retorna uma conta específica
func (h *accountHandle) GetAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			utils.SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		accountID := r.PathValue("id")
		id, err := uuid.Parse(accountID)
		if err != nil {
			h.log.Warn("ID inválido", "error", err)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		if !authAccount.IsAdmin() && authAccount.ID != id {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
			return
		}

		account, err := h.repo.GetByID(id)
		if err != nil {
			h.log.Warn("Conta não encontrada", "account_id", id.String())
			utils.SendError(w, http.StatusNotFound, "Conta não encontrada")
			return
		}

		h.log.Debug("Conta encontrada", "account_id", id.String())
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewAccountResponseDTO(account))
	}
}

// UpdateAccountHandler atualiza uma conta
func (h *accountHandle) UpdateAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			utils.SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		accountID := r.PathValue("id")
		id, err := uuid.Parse(accountID)
		if err != nil {
			h.log.Warn("ID inválido", "error", err)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		var updateDTO dto.AccountUpdateDTO
		if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
			h.log.Warn("Erro ao decodificar JSON", "error", err)
			utils.SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		if !authAccount.IsAdmin() && authAccount.ID != id {
			h.log.Warn("Apenas administradores podem atualizar outras contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem atualizar outras contas")
			return
		}

		if err := updateDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação", "error", err.Error())
			utils.SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		updateData, _ := json.Marshal(updateDTO)
		updatedAccount, err := h.repo.UpdateByID(id, updateData)
		if err != nil {
			h.log.Error("Erro ao atualizar conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao atualizar conta")
			return
		}

		h.log.Info("Conta atualizada com sucesso", "account_id", updatedAccount.ID.String())
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(dto.NewAccountResponseDTO(updatedAccount))
	}
}

// DeleteAccountHandler remove uma conta
func (h *accountHandle) DeleteAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			utils.SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		accountID := r.PathValue("id")
		id, err := uuid.Parse(accountID)
		if err != nil {
			h.log.Warn("ID inválido", "error", err)
			utils.SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		if !authAccount.IsAdmin() {
			h.log.Warn("Apenas administradores podem deletar contas")
			utils.SendError(w, http.StatusForbidden, "Apenas administradores podem deletar contas")
			return
		}

		err = h.repo.DeleteByID(id)
		if err != nil {
			h.log.Error("Erro ao deletar conta", "error", err)
			utils.SendError(w, http.StatusInternalServerError, "Erro ao deletar conta")
			return
		}

		h.log.Info("Conta deletada", "account_id", id.String())
		w.WriteHeader(http.StatusOK)
		w.WriteHeader(http.StatusNoContent)
	}
}
