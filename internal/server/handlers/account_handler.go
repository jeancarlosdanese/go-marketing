// File: internal/server/handlers/account_handler.go

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
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
		var account models.Account
		// 🔥 Decodificar JSON corretamente
		if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
			h.log.Error("Erro ao decodificar JSON", "error", err)
			SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// 🔥 Validar os dados antes de criar a conta
		if err := account.Validate(false); err != nil {
			h.log.Warn("Erro de validação: " + err.Error())
			SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// 🔥 Criar a conta
		createdAccount, err := h.repo.Create(&account)
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

		h.log.Info(fmt.Sprintf("✅ Conta criada com sucesso: %v", createdAccount))
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdAccount)
	}
}

// GetAllAccountsHandler retorna todas as contas cadastradas
func (h *accountHandle) GetAllAccountsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔍 Buscar todas as contas
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

		// 🔍 Buscar todas as contas
		accounts, err := h.repo.GetAll()
		if err != nil {
			h.log.Error("Erro ao buscar contas", "error", err)
			SendError(w, http.StatusInternalServerError, "Erro ao buscar contas")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accounts)
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

		// 🔥 Extrair o ID da URL
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			h.log.Error("ID não fornecido")
			SendError(w, http.StatusBadRequest, "ID não fornecido")
			return
		}

		// 🔥 Se o ID não for um UUID válido, retorna erro
		accountID, err := uuid.Parse(pathParts[2])
		if err != nil {
			h.log.Error("ID inválido", "error", err)
			SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// Apenas administradores podem buscar outras contas
		if !authAccount.IsAdmin() && authAccount.ID != accountID {
			h.log.Warn("Apenas administradores podem buscar outras contas")
			SendError(w, http.StatusForbidden, "Apenas administradores podem buscar outras contas")
			return
		}

		// 🔍 Buscar a conta pelo ID
		account, err := h.repo.GetByID(accountID)
		if err != nil {
			h.log.Error("Erro ao buscar conta", "error", err)
			SendError(w, http.StatusNotFound, "Conta não encontrada")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
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

		// 🔥 Extrair o ID da URL
		pathParts := strings.Split(r.URL.Path, "/")

		// Se não houver ID, retorna erro
		if len(pathParts) < 3 {
			h.log.Error("ID não fornecido")
			SendError(w, http.StatusBadRequest, "ID não fornecido")
			return
		}

		// Se o ID não for um UUID válido, retorna erro
		accountID, err := uuid.Parse(pathParts[2])
		if err != nil {
			h.log.Error("ID inválido", "error", err)
			SendError(w, http.StatusBadRequest, "ID inválido")
			return
		}

		// 🔥 Lendo JSON corretamente do corpo da requisição
		jsonData, err := io.ReadAll(r.Body)
		if err != nil {
			h.log.Error("Erro ao ler corpo da requisição", "error", err)
			SendError(w, http.StatusBadRequest, "Erro ao processar requisição")
			return
		}
		defer r.Body.Close()

		// Se o JSON estiver vazio, retorna erro
		if len(jsonData) == 0 {
			h.log.Error("JSON vazio")
			SendError(w, http.StatusBadRequest, "JSON vazio")
			return
		}

		// Apenas administradores podem atualizar outras contas
		if !authAccount.IsAdmin() && authAccount.ID != accountID {
			h.log.Warn("Apenas administradores podem atualizar outras contas")
			SendError(w, http.StatusForbidden, "Apenas administradores podem atualizar outras contas")
			return
		}

		// 🔍 Buscar a conta antes da atualização
		_, err = h.repo.GetByID(accountID)
		if err != nil {
			h.log.Error("Erro ao buscar conta", "error", err)
			SendError(w, http.StatusNotFound, "Conta não encontrada")
			return
		}

		// Apenas administradores podem alterar e-mail e WhatsApp
		if !authAccount.IsAdmin() {
			// Verifica se o JSON contém campos proibidos
			var updateData map[string]interface{}
			if err := json.Unmarshal(jsonData, &updateData); err != nil {
				h.log.Error("Erro ao decodificar JSON", "error", err)
				SendError(w, http.StatusBadRequest, "Erro ao decodificar JSON")
				return
			}

			// Proibir alteração de e-mail e WhatsApp
			if _, exists := updateData["email"]; exists {
				h.log.Warn("Apenas administradores podem alterar o e-mail")
				SendError(w, http.StatusForbidden, "Apenas administradores podem alterar o e-mail")
				return
			}
			if _, exists := updateData["whatsapp"]; exists {
				h.log.Warn("Apenas administradores podem alterar o WhatsApp")
				SendError(w, http.StatusForbidden, "Apenas administradores podem alterar o WhatsApp")
				return
			}
		}

		// 🔄 Atualizar a conta
		updatedAccount, err := h.repo.UpdateByID(accountID, jsonData)
		if err != nil {
			h.log.Error("Erro ao atualizar conta", "error", err)
			SendError(w, http.StatusInternalServerError, "Erro ao atualizar conta")
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedAccount)
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

		// 🔥 Extrair o ID da URL
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			h.log.Error("ID não fornecido")
			SendError(w, http.StatusBadRequest, "ID não fornecido")
			return
		}

		// 🔥 Se o ID não for um UUID válido, retorna erro
		accountID, err := uuid.Parse(pathParts[2])
		if err != nil {
			h.log.Error("ID inválido", "error", err)
			SendError(w, http.StatusBadRequest, "ID inválido")
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
		if accountID == adminUUID {
			h.log.Warn("O administrador não pode ser deletado")
			SendError(w, http.StatusForbidden, "O administrador não pode ser deletado")
			return
		}

		// ❌ Deleta a conta e retorna o AccountID deletado
		deletedID, err := h.repo.DeleteByID(accountID)
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
