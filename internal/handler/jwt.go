package handler

import (
	"github.com/golang-jwt/jwt/v5"
	"os"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

// Генерация JWT
func generateJWTToken(login string) (string, error) {
    claims := jwt.MapClaims{
        "login": login,
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtKey)
}
