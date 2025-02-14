// File: internal/auth/jwt_service.go

package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("MINHA_CHAVE_SECRETA")

// GenerateJWT cria um token válido por 7 dias
func GenerateJWT(accountID string) (string, error) {
	claims := jwt.MapClaims{
		"account_id": accountID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}

// ValidateJWT verifica se o token é válido
func ValidateJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
}
