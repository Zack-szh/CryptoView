package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/szh/cryptoview/services/api"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("failed to load .env %v", err)
	}

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	server := api.New()
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
