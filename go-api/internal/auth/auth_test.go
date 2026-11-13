package auth

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("test-secret")

func newApp() *fiber.App {
	app := fiber.New()
	app.Get("/p", Middleware(secret), func(c fiber.Ctx) error { return c.SendString("ok") })
	return app
}

func status(t *testing.T, app *fiber.App, authorization string) int {
	t.Helper()
	req := httptest.NewRequest("GET", "/p", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

func TestMiddleware(t *testing.T) {
	app := newApp()
	valid, _ := Issue(secret, "u", time.Hour)
	expired, _ := Issue(secret, "u", -time.Minute)
	otherKey, _ := Issue([]byte("other"), "u", time.Hour)
	hs512, _ := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}).SignedString(secret)
	none, _ := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}).SignedString(jwt.UnsafeAllowNoneSignatureType)

	cases := map[string]struct {
		header string
		want   int
	}{
		"no token":    {"", 401},
		"not bearer":  {"Basic abc", 401},
		"garbage":     {"Bearer not.a.jwt", 401},
		"wrong key":   {"Bearer " + otherKey, 401},
		"expired":     {"Bearer " + expired, 401},
		"alg HS512":   {"Bearer " + hs512, 401},
		"alg none":    {"Bearer " + none, 401},
		"valid HS256": {"Bearer " + valid, 200},
	}
	for name, tc := range cases {
		if got := status(t, app, tc.header); got != tc.want {
			t.Errorf("%s: status %d, want %d", name, got, tc.want)
		}
	}
}

func TestCredentialsMatch(t *testing.T) {
	if !CredentialsMatch("admin", "pw", "admin", "pw") {
		t.Error("expected match")
	}
	if CredentialsMatch("admin", "nope", "admin", "pw") || CredentialsMatch("x", "pw", "admin", "pw") {
		t.Error("expected mismatch")
	}
	if CredentialsMatch("", "", "", "") {
		t.Error("empty configured credentials must never match")
	}
}
