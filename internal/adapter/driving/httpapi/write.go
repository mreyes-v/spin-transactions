package httpapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"spin-transactions/internal/application"
	tx "spin-transactions/internal/domain/transaction"
)

type createRequest struct {
	AccountID   string  `json:"accountId" binding:"required"`
	Type        string  `json:"type" binding:"required"`
	Amount      float64 `json:"amount" binding:"required"`
	Currency    string  `json:"currency" binding:"required"`
	Description string  `json:"description" binding:"required"`
}

// WriteHandlers is the driving adapter for POST /transactions.
type WriteHandlers struct {
	execution *application.WriteService
}

func NewWriteHandlers(execution *application.WriteService) *WriteHandlers {
	return &WriteHandlers{execution: execution}
}

func NewWriteRouter(h *WriteHandlers) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.POST("/transactions", h.Create)
	return r
}

func (h *WriteHandlers) Create(c *gin.Context) {
	var req createRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("error create transaction: invalid json: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	out, err := h.execution.Create(c.Request.Context(), tx.CreateTransaction{
		AccountID:   req.AccountID,
		Type:        tx.Type(req.Type),
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
	})
	if err != nil {
		var validationErr *tx.ValidationError
		if errors.As(err, &validationErr) {
			log.Printf("error create transaction: validation: %v", validationErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		var rejectionErr *tx.ProviderRejectionError
		if errors.As(err, &rejectionErr) {
			log.Printf("error create transaction: provider rejected account=%s code=%s: %s", req.AccountID, rejectionErr.Code, rejectionErr.Message)
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"transaction": toResp(rejectionErr.Transaction),
				"code":        rejectionErr.Code,
				"message":     rejectionErr.Message,
			})
			return
		}
		log.Printf("error create transaction: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, toResp(out))
}
