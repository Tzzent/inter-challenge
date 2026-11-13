// Command server ejecuta la API QR en Go.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"go-api/internal/handler"
	"go-api/internal/statsclient"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "GET this server's /health and exit 0 on 200 (for container healthchecks)")
	flag.Parse()

	port := getenv("PORT", "3000")
	if *healthcheck {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get("http://127.0.0.1:" + port + "/health")
		if err != nil || resp.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		os.Exit(0)
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	app := handler.NewApp(handler.Config{
		Stats:     statsclient.New(getenv("STATS_API_URL", "http://localhost:4000")),
		JWTSecret: []byte(secret),
		Username:  os.Getenv("AUTH_USERNAME"),
		Password:  os.Getenv("AUTH_PASSWORD"),
		WebDir:    getenv("WEB_DIR", "./web"),
	})
	log.Fatal(app.Listen(":" + port))
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
