package httpapi

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"spin-transactions/internal/application"
	tx "spin-transactions/internal/domain/transaction"
)

const (
	defaultPage  = 1
	defaultLimit = 20
	maxLimit     = 100
)

// ReadHandlers is the driving adapter for GET /transactions.
type ReadHandlers struct {
	query *application.ReadService
}

func NewReadHandlers(query *application.ReadService) *ReadHandlers {
	return &ReadHandlers{query: query}
}

func NewReadRouter(h *ReadHandlers) *gin.Engine {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	r.GET("/transactions", h.List)
	return r
}

func (h *ReadHandlers) List(c *gin.Context) {
	page, err := parsePositiveIntQuery(c.Query("page"), defaultPage)
	if err != nil {
		log.Printf("error list transactions: invalid page: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "page must be a positive integer"})
		return
	}
	limit, err := parsePositiveIntQuery(c.Query("limit"), defaultLimit)
	if err != nil {
		log.Printf("error list transactions: invalid limit: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer"})
		return
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	txs, total, err := h.query.List(c.Request.Context(), tx.ListFilter{
		AccountID: c.Query("accountId"),
		Status:    tx.Status(c.Query("status")),
		Type:      tx.Type(c.Query("type")),
		Page:      page,
		Limit:     limit,
	})
	if err != nil {
		log.Printf("error list transactions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	data := make([]transactionResponse, 0, len(txs))
	for _, t := range txs {
		data = append(data, toResp(t))
	}
	c.JSON(http.StatusOK, gin.H{"data": data, "page": page, "limit": limit, "total": total})
}

func parsePositiveIntQuery(raw string, defaultValue int) (int, error) {
	if raw == "" {
		return defaultValue, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 1 {
		return 0, strconv.ErrSyntax
	}
	return value, nil
}
