package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"os"
	"time"

	"featureflag-api/internal/handlers"
	"featureflag-api/internal/middleware"
	"featureflag-api/internal/store"
)

const ServiceVersion = "1.0.0"

func main() {
	apiKey := os.Getenv("FLAG_API_KEY")
	if apiKey == "" {
		buf := make([]byte, 32)
		if _, err := rand.Read(buf); err != nil {
			log.Fatalf("failed to generate API key: %v", err)
		}
		apiKey = hex.EncodeToString(buf)
		log.Printf("WARNING: FLAG_API_KEY not set; generated ephemeral key %s", apiKey)
	}

	handlers.ServiceVersion = ServiceVersion

	s := store.New()
	mux := http.NewServeMux()
	handlers.Register(mux, s)

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8080"
	}

	server := &http.Server{
		Addr:           addr,
		Handler:        middleware.Logging(middleware.Auth(mux, apiKey)),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
