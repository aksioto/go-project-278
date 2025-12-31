package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPing(t *testing.T) {
	router := newRouter()

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d, want %d", w.Code, http.StatusOK)
	}

	if w.Body.String() != "pong" {
		t.Fatalf("unexpected body: got %q, want %q", w.Body.String(), "pong")
	}
}
