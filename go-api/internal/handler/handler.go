// Package handler define las rutas HTTP de la API Go.
package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/static"

	"go-api/internal/auth"
	"go-api/internal/qr"
	"go-api/internal/statsclient"
)

// MaxDim es el máximo de filas y de columnas que acepta POST /api/v1/qr.
const MaxDim = 100

// Config agrupa las dependencias de la app HTTP.
type Config struct {
	Stats     *statsclient.Client
	JWTSecret []byte
	Username  string
	Password  string
	WebDir    string // opcional; si se define, se sirve como estático
}

// NewApp construye la app Fiber con todas las rutas.
func NewApp(cfg Config) *fiber.App {
	app := fiber.New(fiber.Config{
		BodyLimit: 1 << 20, // 1 MiB: una matriz 100x100 cabe con holgura
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			msg := "internal error"
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code, msg = fe.Code, fe.Message
			}
			return c.Status(code).JSON(fiber.Map{"error": msg})
		},
	})

	app.Get("/health", func(c fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		statsAPI := "ok"
		if err := cfg.Stats.Ping(ctx); err != nil {
			statsAPI = "down"
		}
		return c.JSON(fiber.Map{"status": "ok", "statsApi": statsAPI})
	})

	app.Post("/auth/token", func(c fiber.Ctx) error {
		var req struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return badRequest(c, "invalid JSON body")
		}
		if !auth.CredentialsMatch(req.Username, req.Password, cfg.Username, cfg.Password) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid credentials"})
		}
		token, err := auth.Issue(cfg.JWTSecret, req.Username, time.Hour)
		if err != nil {
			return err
		}
		return c.JSON(fiber.Map{"token": token})
	})

	api := app.Group("/api/v1", auth.Middleware(cfg.JWTSecret))
	api.Post("/qr", func(c fiber.Ctx) error {
		var req struct {
			Matrix [][]float64 `json:"matrix"`
		}
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return badRequest(c, "invalid JSON body: matrix must be an array of arrays of numbers")
		}
		if msg := validate(req.Matrix); msg != "" {
			return badRequest(c, msg)
		}
		q, r := qr.Decompose(req.Matrix)
		stats, err := cfg.Stats.Compute(c.Context(), c.Get(fiber.HeaderAuthorization), q, r)
		if err != nil {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "stats service unavailable"})
		}
		return c.JSON(fiber.Map{"q": q, "r": r, "stats": stats})
	})

	if cfg.WebDir != "" {
		app.Get("/*", static.New(cfg.WebDir))
	}
	return app
}

func badRequest(c fiber.Ctx, msg string) error {
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": msg})
}

// validate devuelve un mensaje de error, o "" si la matriz es no vacía, rectangular, finita y cabe en MaxDim.
func validate(m [][]float64) string {
	if len(m) == 0 || len(m[0]) == 0 {
		return "matrix must be a non-empty array of non-empty rows"
	}
	if len(m) > MaxDim || len(m[0]) > MaxDim {
		return fmt.Sprintf("matrix must be at most %dx%d", MaxDim, MaxDim)
	}
	for i, row := range m {
		if len(row) != len(m[0]) {
			return fmt.Sprintf("row %d has %d columns, expected %d (all rows must have the same length)", i, len(row), len(m[0]))
		}
		for _, x := range row {
			if math.IsNaN(x) || math.IsInf(x, 0) {
				return "matrix values must be finite numbers"
			}
		}
	}
	return ""
}
