package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"

	"spin-transactions/internal/bootstrap"
)

func main() {
	_ = godotenv.Load()

	router, err := bootstrap.NewReadRouter()
	if err != nil {
		log.Fatal(err)
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	log.Printf("read server listening on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("read server failed: %v", err)
	}
}
