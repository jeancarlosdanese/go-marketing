// File: /internal/auth/turnstile.go

package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes"`
	ChallengeTS string   `json:"challenge_ts"`
}

func VerifyTurnstile(token string) bool {
	log := logger.GetLogger()

	secretKey := os.Getenv("TURNSTILE_SECRET_KEY")
	if secretKey == "" {
		fmt.Println("TURNSTILE_SECRET_KEY não está configurada.")
		return false
	}

	url := "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	data := fmt.Sprintf("secret=%s&response=%s", secretKey, token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Post(url, "application/x-www-form-urlencoded", bytes.NewBuffer([]byte(data)))

	if err != nil {
		log.Error("Erro ao verificar Turnstile", "error", err)
		return false
	}
	defer resp.Body.Close()

	var result TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Erro ao decodificar resposta do Turnstile:", err)
		return false
	}

	return result.Success
}
