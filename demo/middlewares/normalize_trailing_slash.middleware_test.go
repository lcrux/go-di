package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeTrailingSlashMiddleware(t *testing.T) {
	var got string
	h := NormalizeTrailingSlashMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/todos/", nil)
	h.ServeHTTP(rec, req)

	if got != "/api/todos" {
		t.Fatalf("expected /api/todos, got %q", got)
	}
}
