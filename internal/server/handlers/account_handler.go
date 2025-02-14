// File: internal/server/handlers/account_handler.go

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
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
	return &accountHandle{repo}
}

// CreateAccountHandler retorna uma função que lida com criação de contas
func (h *accountHandle) CreateAccountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var account models.Account
		if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
			http.Error(w, "Erro ao decodificar JSON", http.StatusBadRequest)
			return
		}

		accountCreated, err := h.repo.Create(&account)
		if err != nil {
			http.Error(w, "Erro ao criar conta", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(accountCreated)
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
