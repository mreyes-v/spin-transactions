package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"spin-transactions/internal/adapter/driving/httpapi"
	"spin-transactions/internal/application"
	"spin-transactions/internal/domain/events"
	tx "spin-transactions/internal/domain/transaction"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

type stubProvider struct{ res tx.ProviderResult }

func (s stubProvider) Execute(_ context.Context, _ tx.ProviderExecuteRequest) (tx.ProviderResult, error) {
	return s.res, nil
}

type stubRepo struct{ bal float64 }

func (s stubRepo) Save(_ context.Context, _ tx.Transaction) error { return nil }
func (s stubRepo) CurrentBalance(_ context.Context, _ string) (float64, error) {
	return s.bal, nil
}
func (s stubRepo) List(_ context.Context, _ tx.ListFilter) ([]tx.Transaction, int, error) {
	return []tx.Transaction{{ID: "1"}}, 1, nil
}

type stubPub struct{}

func (stubPub) Publish(_ context.Context, _ events.DomainEvent) error { return nil }

func TestWriteHandler_Create(t *testing.T) {
	svc := application.NewWriteService(
		stubProvider{res: tx.ProviderResult{TransactionID: "x", Status: tx.ProviderStatusApproved}},
		stubRepo{bal: 0},
		stubPub{},
	)
	h := httpapi.NewWriteHandlers(svc)
	body, _ := json.Marshal(map[string]any{
		"accountId": "a", "type": "CREDIT", "amount": 10.0, "currency": "MXN", "description": "d",
	})
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/transactions", bytes.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	h.Create(c)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
}

func TestReadHandler_List(t *testing.T) {
	h := httpapi.NewReadHandlers(application.NewReadService(stubRepo{}))
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/transactions", nil)
	h.List(c)
	if rec.Code != http.StatusOK {
		t.Fatalf("list %d", rec.Code)
	}
}

func TestRouters_SplitRoutes(t *testing.T) {
	writeRouter := httpapi.NewWriteRouter(httpapi.NewWriteHandlers(application.NewWriteService(
		stubProvider{res: tx.ProviderResult{Status: tx.ProviderStatusApproved}},
		stubRepo{},
		stubPub{},
	)))
	readRouter := httpapi.NewReadRouter(httpapi.NewReadHandlers(application.NewReadService(stubRepo{})))

	body, _ := json.Marshal(map[string]any{
		"accountId": "a", "type": "CREDIT", "amount": 10.0, "currency": "MXN", "description": "d",
	})

	assertStatus(t, writeRouter, http.MethodGet, "/health", nil, http.StatusOK)
	assertStatus(t, readRouter, http.MethodGet, "/health", nil, http.StatusOK)
	assertStatus(t, writeRouter, http.MethodGet, "/transactions", nil, http.StatusNotFound)
	assertStatus(t, readRouter, http.MethodPost, "/transactions", body, http.StatusNotFound)
	assertStatus(t, readRouter, http.MethodGet, "/transactions", nil, http.StatusOK)
}

func assertStatus(t *testing.T, h http.Handler, method, path string, body []byte, want int) {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	h.ServeHTTP(rec, req)
	if rec.Code != want {
		t.Fatalf("%s %s = %d, want %d (%s)", method, path, rec.Code, want, rec.Body.String())
	}
}
