// File: internal/auth/auth_service.go

package auth

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
)

// GenerateOTP cria uma senha numérica de 8 dígitos
func GenerateOTP() (string, error) {
	max := big.NewInt(90000000) // Intervalo máximo (99999999 - 10000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%08d", n.Int64()+10000000), nil
}

// SendOTP envia o OTP para o e-mail ou WhatsApp
func SendOTP(destination string, otp string) {
	// Aqui podemos integrar com um serviço de e-mail ou WhatsApp
	log.Printf("📩 Enviando OTP para %s: %s", destination, otp)
}
