package order

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerServeHTTP(t *testing.T) {
	h := NewHandler(NewService())
	req := httptest.NewRequest(http.MethodGet, "/orders", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
