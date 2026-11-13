package handler

import (
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"

	"go-api/internal/auth"
	"go-api/internal/statsclient"
)

var secret = []byte("test-secret")

// fakeNode stands in for the Node stats API and records the forwarded Authorization header.
func fakeNode(t *testing.T, status int) (*httptest.Server, *string) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			return
		}
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(status)
		w.Write([]byte(`{"max":3,"min":-5,"average":-0.95,"sum":-7.6,"anyDiagonal":false,"diagonal":[false,false]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &gotAuth
}

func newApp(statsURL string) *fiber.App {
	return NewApp(Config{Stats: statsclient.New(statsURL), JWTSecret: secret, Username: "admin", Password: "pw"})
}

func do(t *testing.T, app *fiber.App, method, path, body, authorization string) (int, map[string]any) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 10 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	json.Unmarshal(raw, &out)
	return resp.StatusCode, out
}

func bearer(t *testing.T) string {
	tok, err := auth.Issue(secret, "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return "Bearer " + tok
}

func TestQRSuccessForwardsAuthAndReturnsStats(t *testing.T) {
	node, gotAuth := fakeNode(t, 200)
	app := newApp(node.URL)
	authz := bearer(t)
	code, body := do(t, app, "POST", "/api/v1/qr", `{"matrix":[[3,0],[4,5]]}`, authz)
	if code != 200 {
		t.Fatalf("status %d, body %v", code, body)
	}
	if *gotAuth != authz {
		t.Errorf("Authorization not forwarded: %q", *gotAuth)
	}
	q := body["q"].([]any)
	if v := q[0].([]any)[0].(float64); math.Abs(v - -0.6) > 1e-9 {
		t.Errorf("Q[0][0] = %v", v)
	}
	if body["stats"].(map[string]any)["sum"].(float64) != -7.6 {
		t.Errorf("stats not included: %v", body["stats"])
	}
}

func TestQRValidation(t *testing.T) {
	node, _ := fakeNode(t, 200)
	app := newApp(node.URL)
	authz := bearer(t)
	big := "[" + strings.Repeat("[1],", 100) + "[1]]"
	wide := "[[" + strings.Repeat("1,", 100) + "1]]"
	cases := map[string]string{
		"malformed JSON": `{"matrix":[[1,2]`,
		"missing matrix": `{}`,
		"empty matrix":   `{"matrix":[]}`,
		"empty row":      `{"matrix":[[]]}`,
		"ragged rows":    `{"matrix":[[1,2],[3]]}`,
		"null row":       `{"matrix":[[1],null]}`,
		"non-numeric":    `{"matrix":[[1,"a"]]}`,
		"non-finite":     `{"matrix":[[1e400]]}`,
		"too many rows":  `{"matrix":` + big + `}`,
		"too many cols":  `{"matrix":` + wide + `}`,
	}
	for name, payload := range cases {
		code, body := do(t, app, "POST", "/api/v1/qr", payload, authz)
		if code != 400 || body["error"] == nil {
			t.Errorf("%s: status %d body %v, want 400 with error", name, code, body)
		}
	}
}

func TestQRAuth(t *testing.T) {
	node, _ := fakeNode(t, 200)
	app := newApp(node.URL)
	if code, _ := do(t, app, "POST", "/api/v1/qr", `{"matrix":[[1]]}`, ""); code != 401 {
		t.Errorf("no token: %d", code)
	}
	if code, _ := do(t, app, "POST", "/api/v1/qr", `{"matrix":[[1]]}`, "Bearer bad"); code != 401 {
		t.Errorf("bad token: %d", code)
	}
	if code, _ := do(t, app, "POST", "/api/v1/qr", `{"matrix":[[1]]}`, bearer(t)); code != 200 {
		t.Errorf("valid token: %d", code)
	}
}

func TestQRStatsFailureIs502(t *testing.T) {
	for _, status := range []int{500, 401} {
		node, _ := fakeNode(t, status)
		code, body := do(t, newApp(node.URL), "POST", "/api/v1/qr", `{"matrix":[[1]]}`, bearer(t))
		if code != 502 || body["error"] == nil {
			t.Errorf("node %d: status %d body %v, want 502", status, code, body)
		}
	}
	code, _ := do(t, newApp("http://127.0.0.1:1"), "POST", "/api/v1/qr", `{"matrix":[[1]]}`, bearer(t))
	if code != 502 {
		t.Errorf("node down: status %d, want 502", code)
	}
}

func TestHealth(t *testing.T) {
	node, _ := fakeNode(t, 200)
	if code, body := do(t, newApp(node.URL), "GET", "/health", "", ""); code != 200 || body["statsApi"] != "ok" {
		t.Errorf("node up: %d %v", code, body)
	}
	if code, body := do(t, newApp("http://127.0.0.1:1"), "GET", "/health", "", ""); code != 200 || body["statsApi"] != "down" {
		t.Errorf("node down: %d %v", code, body)
	}
}

func TestAuthToken(t *testing.T) {
	app := newApp("http://127.0.0.1:1")
	code, body := do(t, app, "POST", "/auth/token", `{"username":"admin","password":"pw"}`, "")
	if code != 200 || body["token"] == nil {
		t.Fatalf("valid creds: %d %v", code, body)
	}
	if code, _ := do(t, app, "POST", "/auth/token", `{"username":"admin","password":"nope"}`, ""); code != 401 {
		t.Errorf("bad creds: %d", code)
	}
	if code, _ := do(t, app, "POST", "/auth/token", `not json`, ""); code != 400 {
		t.Errorf("malformed: %d", code)
	}
}
