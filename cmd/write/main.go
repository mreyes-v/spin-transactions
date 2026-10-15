package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"spin-transactions/internal/bootstrap"
)

func main() {
	_ = godotenv.Load()

	router, err := bootstrap.NewWriteRouter()
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	log.Printf("write server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("write server failed: %v", err)
	}
}
