// Package auth emite y verifica JWT HS256.
package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Issue firma un token HS256 para subject que expira tras ttl.
func Issue(secret []byte, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// CredentialsMatch compara en tiempo constante; el hash previo oculta diferencias de longitud.
// Credenciales esperadas vacías nunca coinciden: un servidor sin configurar rechaza todo login.
func CredentialsMatch(user, pass, wantUser, wantPass string) bool {
	if wantUser == "" || wantPass == "" {
		return false
	}
	u1, u2 := sha256.Sum256([]byte(user)), sha256.Sum256([]byte(wantUser))
	p1, p2 := sha256.Sum256([]byte(pass)), sha256.Sum256([]byte(wantPass))
	return subtle.ConstantTimeCompare(u1[:], u2[:])&subtle.ConstantTimeCompare(p1[:], p2[:]) == 1
}

// Middleware rechaza peticiones sin un "Bearer <jwt>" válido firmado con secret.
// Solo acepta HS256 (bloquea alg=none y confusión de algoritmo) y exige expiración.
func Middleware(secret []byte) fiber.Handler {
	return func(c fiber.Ctx) error {
		token, ok := strings.CutPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
		if !ok || token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing bearer token"})
		}
		_, err := jwt.Parse(token, func(*jwt.Token) (any, error) { return secret, nil },
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}
		return c.Next()
	}
}
