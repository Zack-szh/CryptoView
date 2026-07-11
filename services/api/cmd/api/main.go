package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/szh/cryptoview/services/api"
	"github.com/szh/cryptoview/services/api/db"
)

func main() {
	// build context for signal-aware exit
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := godotenv.Load()

	if err != nil {
		log.Fatalf("failed to load .env %v", err)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatalf("DATABASE_URL is not set")
	}

	// init db connection, note that this one is a separate connection from market-data/cmd/stream/main.go
	store, err := db.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	server := api.New(store)
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
