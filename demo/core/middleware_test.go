package core

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChain(t *testing.T) {
	order := []string{}
	mw1 := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1")
			next.ServeHTTP(w, r)
		}
	}
	mw2 := func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2")
			next.ServeHTTP(w, r)
		}
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		order = append(order, "final")
		w.WriteHeader(http.StatusOK)
	})

	chained := Chain(mw1, mw2)
	rec := httptest.NewRecorder()
	chained(final).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if len(order) != 3 || order[0] != "mw1" || order[1] != "mw2" || order[2] != "final" {
		t.Fatalf("unexpected order: %v", order)
	}
}
