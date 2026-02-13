package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCleanUpTrailingSlash(t *testing.T) {
	if got := CleanUpTrailingSlash(" /api/ "); got != "/api" {
		t.Fatalf("expected /api, got %q", got)
	}
	if got := CleanUpTrailingSlash("/"); got != "/" {
		t.Fatalf("expected /, got %q", got)
	}
}

func TestJoinUrlPath(t *testing.T) {
	got, err := JoinURLPath("/api/", "/todos/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "/api/todos" {
		t.Fatalf("expected /api/todos, got %q", got)
	}
}

func TestServerMuxRouter_GroupAndAddGet(t *testing.T) {
	router := NewServerMuxRouter()
	api := router.Group("api")

	called := false
	handler := func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	if err := api.AddGet("/todos", handler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/todos", nil)
	rec := httptest.NewRecorder()
	router.Handler().ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
