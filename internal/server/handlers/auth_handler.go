// File: /internal/server/handlers/auth_handler.go

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/auth"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
)

type AuthHandle interface {
	RequestAuthHandle() http.HandlerFunc
	VerifyAuthHandle() http.HandlerFunc
}

type authHandle struct {
	repo db.AccountOTPRepository
}

func NewAuthHandle(repo db.AccountOTPRepository) AuthHandle {
	return &authHandle{repo}
}

func (h *authHandle) RequestAuthHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Identifier string `json:"identifier"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erro ao processar JSON", http.StatusBadRequest)
			return
		}

		account, err := h.repo.FindByEmailOrWhatsApp(req.Identifier)
		if err != nil {
			http.Error(w, "Usuário não encontrado", http.StatusNotFound)
			return
		}

		// Gerar e enviar OTP...
		otp, _ := auth.GenerateOTP()
		h.repo.StoreOTP(account.ID.String(), otp)
		auth.SendOTP(req.Identifier, otp)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "OTP enviado"})
	}
}

func (h *authHandle) VerifyAuthHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Identifier string `json:"identifier"`
			OTP        string `json:"otp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Erro ao processar JSON", http.StatusBadRequest)
			return
		}

		// 🔥 Limpar OTPs expirados antes de verificar
		h.repo.CleanExpiredOTPs()

		// 🔥 Buscar OTP válido
		accountID, err := h.repo.FindValidOTP(req.Identifier, req.OTP)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// 🔥 Gerar token JWT
		token, _ := auth.GenerateJWT(accountID.String())

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}
