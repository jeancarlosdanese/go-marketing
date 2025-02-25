// File: internal/auth/jwt_service.go

package auth

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jeancarlosdanese/go-marketing/internal/models"
)

var secretKey = []byte(os.Getenv("JWT_SECRET_KEY"))

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

// IsAdminByToken verifica se o usuário autenticado é admin
func IsAdminByToken(account *models.Account) bool {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminWhatsApp := os.Getenv("ADMIN_WHATSAPP")

	// Se o ID no token corresponder ao admin, retorna true
	return account.Email == adminEmail || account.WhatsApp == adminWhatsApp
}

// GetAccountIDFromToken extrai o ID da conta do token JWT
func GetAccountIDFromToken(r *http.Request) (string, error) {
	tokenStr := ExtractTokenFromHeader(r)
	if tokenStr == "" {
		return "", errors.New("token não encontrado")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})
	if err != nil {
		return "", errors.New("token inválido")
	}

	// Pegar o `account_id` do token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if accountID, exists := claims["account_id"].(string); exists {
			return accountID, nil
		}
	}
	return "", errors.New("ID da conta não encontrado no token")
}

// ExtractTokenFromHeader pega o token do header Authorization
func ExtractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}
	return parts[1]
}
