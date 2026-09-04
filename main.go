package main

import (
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
		log.Fatal("FLAG_API_KEY must be set")
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
