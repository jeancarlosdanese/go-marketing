// File: internal/server/handlers/account_handler.go

package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/auth"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
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
	repo db.AccountRepository
}

func NewAccountHandle(repo db.AccountRepository) AccountHandle {
	return &accountHandle{repo: repo}
}

// CreateAccountHandler cria uma nova conta
func (h *accountHandle) CreateAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.GetLogger()
		var account models.Account

		// 🔥 Decodificar JSON corretamente
		if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
			log.Warn("Erro ao decodificar JSON: " + err.Error())
			SendError(w, http.StatusBadRequest, "Erro ao decodificar JSON")
			return
		}
		defer r.Body.Close()

		// 🔥 Validar os dados antes de criar a conta
		if err := account.Validate(false); err != nil {
			log.Warn("Erro de validação: " + err.Error())
			SendError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Criar a conta
		createdAccount, err := h.repo.Create(&account)
		if err != nil {
			log.Error("Erro ao criar conta: " + err.Error())

			// 🔥 Tratar erro de chave duplicada no PostgreSQL
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				SendError(w, http.StatusConflict, "E-mail ou WhatsApp já cadastrado")
			} else {
				SendError(w, http.StatusInternalServerError, "Erro ao criar conta")
			}
			return
		}

		log.Info(fmt.Sprintf("✅ Conta criada com sucesso: %v", createdAccount))
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(createdAccount)
	}
}

// GetAllAccountsHandler retorna todas as contas cadastradas
func (h *accountHandle) GetAllAccountsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accounts, err := h.repo.GetAll()
		if err != nil {
			http.Error(w, "Erro ao buscar contas", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accounts)
	}
}

// GetAccountHandler retorna uma função que lida com busca de contas
func (h *accountHandle) GetAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "ID não fornecido", http.StatusBadRequest)
			return
		}
		accountID, err := uuid.Parse(pathParts[2])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		account, err := h.repo.GetByID(accountID)
		if err != nil {
			http.Error(w, "Conta não encontrada", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

// UpdateAccountHandler atualiza uma conta pelo ID
func (h *accountHandle) UpdateAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 🔥 Extrair o ID da URL
		pathParts := strings.Split(r.URL.Path, "/")

		// Se não houver ID, retorna erro
		if len(pathParts) < 3 {
			http.Error(w, "ID não fornecido", http.StatusBadRequest)
			return
		}

		// Se o ID não for um UUID válido, retorna erro
		accountID, err := uuid.Parse(pathParts[2])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		// 🔥 Lendo JSON corretamente do corpo da requisição
		jsonData, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Erro ao ler o corpo da requisição", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		// Se o JSON estiver vazio, retorna erro
		if len(jsonData) == 0 {
			http.Error(w, "Nenhum dado para atualizar", http.StatusBadRequest)
			return
		}

		// 🔍 Buscar a conta antes da atualização
		existingAccount, err := h.repo.GetByID(accountID)
		if err != nil {
			http.Error(w, "Conta não encontrada", http.StatusNotFound)
			return
		}

		// 🔒 Validar se é admin antes de permitir alteração de e-mail ou WhatsApp
		if !auth.IsAdmin(existingAccount) {
			// Verifica se o JSON contém campos proibidos
			var updateData map[string]interface{}
			if err := json.Unmarshal(jsonData, &updateData); err != nil {
				http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
				return
			}

			// Proibir alteração de e-mail e WhatsApp
			if _, exists := updateData["email"]; exists {
				http.Error(w, "Apenas administradores podem alterar o e-mail", http.StatusForbidden)
				return
			}
			if _, exists := updateData["whatsapp"]; exists {
				http.Error(w, "Apenas administradores podem alterar o WhatsApp", http.StatusForbidden)
				return
			}
		}

		// 🔄 Atualizar a conta
		updatedAccount, err := h.repo.UpdateByID(accountID, jsonData)
		if err != nil {
			http.Error(w, "Erro ao atualizar conta", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updatedAccount)
	}
}

// DeleteAccountHandler remove uma conta pelo ID
func (h *accountHandle) DeleteAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "ID não fornecido", http.StatusBadRequest)
			return
		}
		accountID, err := uuid.Parse(pathParts[2])
		if err != nil {
			http.Error(w, "ID inválido", http.StatusBadRequest)
			return
		}

		deletedID, err := h.repo.DeleteByID(accountID)
		if err != nil {
			http.Error(w, "Erro ao excluir conta", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deleted_id": deletedID,
		})
	}
}
