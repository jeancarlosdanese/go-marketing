// File: /internal/auth/recaptcha.go

package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jeancarlosdanese/go-marketing/internal/logger"
)

// Estrutura para resposta do reCAPTCHA
type RecaptchaResponse struct {
	Success     bool     `json:"success"`
	Score       float64  `json:"score"`
	Action      string   `json:"action"`
	ChallengeTs string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
	ErrorCodes  []string `json:"error-codes"`
}

// Verifica se o reCAPTCHA é válido
func VerifyRecaptcha(token string) (bool, error) {
	log := logger.GetLogger()

	secretKey := os.Getenv("RECAPTCHA_SECRET_KEY")
	if secretKey == "" {
		return false, errors.New("reCAPTCHA secret key não configurada")
	}

	url := "https://www.google.com/recaptcha/api/siteverify"
	data := fmt.Sprintf("secret=%s&response=%s", secretKey, token)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(data)))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// ✅ Ler o corpo uma única vez
	body, _ := io.ReadAll(resp.Body)
	log.Debug("Resposta do reCAPTCHA", "body", string(body))

	// ✅ Decodificar a partir da variável `body`
	var recaptchaResponse RecaptchaResponse
	if err := json.Unmarshal(body, &recaptchaResponse); err != nil {
		return false, err
	}

	// ✅ Se for sucesso e tiver um score alto, consideramos válido
	if recaptchaResponse.Success && recaptchaResponse.Score >= 0.5 {
		return true, nil
	}

	// 🚨 Caso falhe, imprimir os erros recebidos
	if !recaptchaResponse.Success {
		log.Debug("Erro na verificação do reCAPTCHA", "error", recaptchaResponse.ErrorCodes)
	}

	return false, errors.New("verificação do reCAPTCHA falhou")
}
