// File: internal/server/handlers/response.go

package handlers

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse estrutura padronizada para erros
type ErrorResponse struct {
	Error string `json:"error"`
}

// SendError retorna uma resposta JSON padronizada
func SendError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{Error: message})
}
