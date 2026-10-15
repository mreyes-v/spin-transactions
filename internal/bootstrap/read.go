package bootstrap

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"spin-transactions/internal/adapter/driven/postgres"
	"spin-transactions/internal/adapter/driving/httpapi"
	"spin-transactions/internal/application"
)

// NewReadRouter wires the read path. It does not call the provider or run migrations.
func NewReadRouter() (*gin.Engine, error) {
	if err := connectDB(false); err != nil {
		return nil, err
	}
	if DB == nil {
		return nil, fmt.Errorf("database is required")
	}

	repo := postgres.NewTransactionRepository(DB)
	svc := application.NewReadService(repo)
	return httpapi.NewReadRouter(httpapi.NewReadHandlers(svc)), nil
}
