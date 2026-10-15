package bootstrap

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"

	"spin-transactions/internal/adapter/driven/events"
	"spin-transactions/internal/adapter/driven/postgres"
	"spin-transactions/internal/adapter/driven/provider"
	"spin-transactions/internal/adapter/driving/httpapi"
	"spin-transactions/internal/application"
)

// NewWriteRouter wires the write path: provider, repository, and publisher.
func NewWriteRouter() (*gin.Engine, error) {
	baseURL := os.Getenv("PROVIDER_BASE_URL")
	apiKey := os.Getenv("PROVIDER_API_KEY")
	if baseURL == "" || apiKey == "" {
		return nil, fmt.Errorf("PROVIDER_BASE_URL and PROVIDER_API_KEY are required")
	}
	if err := connectDB(true); err != nil {
		return nil, err
	}
	if DB == nil {
		return nil, fmt.Errorf("database is required")
	}

	repo := postgres.NewTransactionRepository(DB)
	providerClient := provider.NewHTTPClient(baseURL, apiKey)
	publisher := events.NewMemoryPublisher()
	svc := application.NewWriteService(providerClient, repo, publisher)
	return httpapi.NewWriteRouter(httpapi.NewWriteHandlers(svc)), nil
}
