// File: internal/server/handlers/response.go

// Package handlers é responsável por criar respostas padronizadas para a API
// O pacote funções para criar respostas padronizadas:
// - SendError: envia um erro JSON padronizado
// Para utilizar o pacote, basta importar o pacote e chamar a função desejada.
// Exemplo:
//
//	handlers.SendError(w, http.StatusNotFound, "Recurso não encontrado")

package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse estrutura um erro padronizado para API
type ErrorResponse struct {
	Message string `json:"message"`
	Error   string `json:"error"`
	Code    int    `json:"code"`
}

// SendError retorna um erro JSON padronizado
func SendError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Message: message,
		Error:   http.StatusText(status),
		Code:    status,
	})
}
