// Package auth issues and verifies HS256 JWTs.
package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Issue returns an HS256 token for subject that expires after ttl.
func Issue(secret []byte, subject string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   subject,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(secret)
}

// CredentialsMatch compares credentials in constant time. Hashing first hides length differences.
// Empty expected credentials never match, so an unconfigured server rejects every login.
func CredentialsMatch(user, pass, wantUser, wantPass string) bool {
	if wantUser == "" || wantPass == "" {
		return false
	}
	u1, u2 := sha256.Sum256([]byte(user)), sha256.Sum256([]byte(wantUser))
	p1, p2 := sha256.Sum256([]byte(pass)), sha256.Sum256([]byte(wantPass))
	return subtle.ConstantTimeCompare(u1[:], u2[:])&subtle.ConstantTimeCompare(p1[:], p2[:]) == 1
}

// Middleware rejects requests without a valid "Bearer <jwt>" signed with secret.
// Only HS256 is accepted, so alg=none and algorithm-confusion tokens fail. An expiry is required.
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
