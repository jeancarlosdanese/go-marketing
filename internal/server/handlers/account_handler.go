// File: internal/server/handlers/account_handler.go

package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/dto"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
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

		// 🔥 Decodificar JSON para DTO
		if err := json.NewDecoder(r.Body).Decode(&accountDTO); err != nil {
			h.log.Error("Erro ao decodificar JSON", "error", err)
			SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔥 Validar DTO
		if err := accountDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação: " + err.Error())
			SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// 🔥 Criar a conta no banco
		account := &models.Account{
			Name:     accountDTO.Name,
			Email:    accountDTO.Email,
			WhatsApp: accountDTO.WhatsApp,
		}

		createdAccount, err := h.repo.Create(account)
		if err != nil {
			h.log.Error("Erro ao criar conta", "error", err)

			// 🔥 Tratar erro de chave duplicada no PostgreSQL
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				SendError(w, http.StatusConflict, "E-mail ou WhatsApp já cadastrado")
			} else {
				SendError(w, http.StatusInternalServerError, "Erro ao criar conta")
			}
			return
		}

		// 🔥 Criar resposta DTO já formatada
		response := dto.NewAccountResponseDTO(createdAccount)

		h.log.Info(fmt.Sprintf("✅ Conta criada com sucesso: %v", response))
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
			SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		// Apenas administradores podem buscar todas as contas
		if !authAccount.IsAdmin() {
			h.log.Warn("Apenas administradores podem buscar todas as contas")
			SendError(w, http.StatusForbidden, "Apenas administradores podem buscar todas as contas")
			return
		}

		accounts, err := h.repo.GetAll()
		if err != nil {
			h.log.Error("Erro ao buscar contas", "error", err)
			SendError(w, http.StatusInternalServerError, "Erro ao buscar contas")
			return
		}

		// 🔥 Criar resposta DTO para todas as contas
		var response []dto.AccountResponseDTO
		for _, acc := range accounts {
			// 🔥 Criar resposta DTO já formatada
			accountDto := dto.NewAccountResponseDTO(acc)
			response = append(response, accountDto)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// GetAccountHandler retorna uma função que lida com busca de contas
func (h *accountHandle) GetAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔥 Buscar a conta autenticada no contexto
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		accountID := r.PathValue("id")
		id, err := uuid.Parse(accountID)
		if err != nil {
			h.log.Error("ID inválido", "error", err)
			SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// Apenas administradores podem buscar outras contas
		if !authAccount.IsAdmin() && authAccount.ID != id {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
			return
		}

		// 🔍 Buscar a conta pelo ID
		account, err := h.repo.GetByID(id)
		if err != nil {
			h.log.Error("Erro ao buscar conta", "error", err)
			SendError(w, http.StatusNotFound, "Conta não encontrada")
			return
		}

		accountDto := dto.NewAccountResponseDTO(account)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accountDto)
	}
}

// UpdateAccountHandler atualiza uma conta pelo ID
func (h *accountHandle) UpdateAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		accountID := r.PathValue("id")
		id, err := uuid.Parse(accountID)
		if err != nil {
			h.log.Error("ID inválido", "error", err)
			SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔥 Decodificar JSON para DTO
		var updateDTO dto.AccountUpdateDTO
		if err := json.NewDecoder(r.Body).Decode(&updateDTO); err != nil {
			h.log.Error("Erro ao decodificar JSON", "error", err)
			SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// Apenas administradores podem atualizar outras contas
		if !authAccount.IsAdmin() && authAccount.ID != id {
			h.log.Warn("Apenas administradores podem atualizar outras contas")
			SendError(w, http.StatusForbidden, "Apenas administradores podem atualizar outras contas")
			return
		}

		// Apenas administradores podem alterar e-mail e WhatsApp
		if !authAccount.IsAdmin() {
			if updateDTO.Email != "" || updateDTO.WhatsApp != "" {
				h.log.Warn("Apenas administradores podem alterar e-mail e WhatsApp")
				SendError(w, http.StatusForbidden, "Apenas administradores podem alterar e-mail e WhatsApp")
				return
			}
		}

		// 🔥 Validar DTO
		if err := updateDTO.Validate(); err != nil {
			h.log.Warn("Erro de validação: " + err.Error())
			SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// 🔄 Atualizar a conta
		updateData := map[string]interface{}{}
		if updateDTO.Name != "" {
			updateData["name"] = updateDTO.Name
		}
		if updateDTO.Email != "" {
			updateData["email"] = updateDTO.Email
		}
		if updateDTO.WhatsApp != "" {
			updateData["whatsapp"] = updateDTO.WhatsApp
		}

		jsonData, _ := json.Marshal(updateData)
		updatedAccount, err := h.repo.UpdateByID(id, jsonData)
		if err != nil {
			h.log.Error("Erro ao atualizar conta", "error", err)

			// 🔥 Tratar erro de chave duplicada no PostgreSQL
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				SendError(w, http.StatusConflict, "E-mail ou WhatsApp já cadastrado")
			} else {
				SendError(w, http.StatusInternalServerError, "Erro ao atualizar conta")
			}
			return
		}

		// 🔥 Criar resposta DTO já formatada
		response := dto.NewAccountResponseDTO(updatedAccount)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// DeleteAccountHandler remove uma conta pelo ID
func (h *accountHandle) DeleteAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔥 Buscar a conta autenticada no contexto
		authAccount, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Error("Conta não encontrada no contexto")
			SendError(w, http.StatusInternalServerError, "Conta não encontrada no contexto")
			return
		}

		accountID := r.PathValue("id")
		id, err := uuid.Parse(accountID)
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		// 🔥 Apenas administradores podem deletar contas
		if !authAccount.IsAdmin() {
			h.log.Warn("Apenas administradores podem deletar contas")
			SendError(w, http.StatusForbidden, "Apenas administradores podem deletar contas")
			return
		}

		// 🔥 O administrador não pode ser deletado
		adminUUID, err := uuid.Parse("00000000-0000-0000-0000-000000000001")
		if err != nil {
			h.log.Error("Erro ao parsear UUID do administrador", "error", err)
			SendError(w, http.StatusInternalServerError, "Erro interno do servidor")
			return
		}
		if id == adminUUID {
			h.log.Warn("O administrador não pode ser deletado")
			SendError(w, http.StatusForbidden, "O administrador não pode ser deletado")
			return
		}

		// ❌ Deleta a conta e retorna o AccountID deletado
		deletedID, err := h.repo.DeleteByID(id)
		if err != nil {
			h.log.Error("Erro ao deletar conta", "error", err)
			SendError(w, http.StatusInternalServerError, "Erro ao deletar conta")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deleted_id": deletedID,
		})
	}
}
