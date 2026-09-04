package main

import (
	"log"
	"net/http"
	"time"

	"featureflag-api/internal/handlers"
	"featureflag-api/internal/middleware"
	"featureflag-api/internal/store"
)

func main() {
	s := store.New()
	mux := http.NewServeMux()
	handlers.Register(mux, s)

	server := &http.Server{
		Addr:           ":8080",
		Handler:        middleware.Logging(mux),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
