// File: /internal/server/handlers/auth_handler.go

package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jeancarlosdanese/go-marketing/internal/auth"
	"github.com/jeancarlosdanese/go-marketing/internal/db"
	"github.com/jeancarlosdanese/go-marketing/internal/logger"
	"github.com/jeancarlosdanese/go-marketing/internal/middleware"
)

type AuthHandle interface {
	RequestAuthHandle() http.HandlerFunc
	VerifyAuthHandle() http.HandlerFunc
	MeHandler() http.HandlerFunc
}

type authHandle struct {
	log  *slog.Logger
	repo db.AccountOTPRepository
}

func NewAuthHandle(repo db.AccountOTPRepository) AuthHandle {
	log := logger.GetLogger()
	return &authHandle{log: log, repo: repo}
}

// 🔐 Solicita autenticação (envia OTP)
func (h *authHandle) RequestAuthHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("📩 Recebida solicitação de autenticação")

		var req struct {
			Identifier string `json:"identifier"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.log.Warn("⚠️ Erro ao processar JSON", "error", err)
			http.Error(w, "Erro ao processar JSON", http.StatusBadRequest)
			return
		}

		account, err := h.repo.FindByEmailOrWhatsApp(r.Context(), req.Identifier)
		if err != nil {
			h.log.Warn("⚠️ Usuário não encontrado", "identifier", req.Identifier)
			http.Error(w, "Usuário não encontrado", http.StatusNotFound)
			return
		}

		// Gerar e enviar OTP...
		otp, _ := auth.GenerateOTP()
		h.repo.StoreOTP(r.Context(), account.ID.String(), otp)
		auth.SendOTP(req.Identifier, otp)

		h.log.Info("✅ OTP enviado com sucesso", "identifier", req.Identifier)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "OTP enviado"})
	}
}

// 🔑 Verifica a autenticação (valida OTP e gera JWT)
func (h *authHandle) VerifyAuthHandle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("🔐 Recebida solicitação de verificação de OTP")

		var req struct {
			Identifier string `json:"identifier"`
			OTP        string `json:"otp"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.log.Warn("⚠️ Erro ao processar JSON", "error", err)
			http.Error(w, "Erro ao processar JSON", http.StatusBadRequest)
			return
		}

		// 🔥 Limpar OTPs expirados antes de verificar
		h.repo.CleanExpiredOTPs(r.Context())

		// 🔥 Buscar OTP válido
		accountID, err := h.repo.FindValidOTP(r.Context(), req.Identifier, req.OTP)
		if err != nil {
			h.log.Warn("⚠️ OTP inválido", "identifier", req.Identifier, "error", err)
			http.Error(w, "OTP inválido ou expirado", http.StatusUnauthorized)
			return
		}

		// 🔥 Gerar token JWT
		token, _ := auth.GenerateJWT(accountID.String())

		h.log.Info("✅ OTP validado com sucesso", "identifier", req.Identifier)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

// 📌 Retorna os dados do usuário autenticado
func (h *authHandle) MeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.log.Info("🔎 Obtendo informações do usuário autenticado")

		account, ok := middleware.GetAuthenticatedAccount(r.Context())
		if !ok {
			h.log.Warn("⚠️ Tentativa de acesso não autorizado")
			http.Error(w, "Não autorizado", http.StatusUnauthorized)
			return
		}

		h.log.Info("✅ Dados do usuário retornados", "email", account.Email, "name", account.Name)

		json.NewEncoder(w).Encode(map[string]string{
			"email": account.Email,
			"name":  account.Name,
		})
	}
}
