package testutil

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expected string) {
	t.Helper()

	body := DecodeJSON[map[string]string](t, resp)
	if !strings.Contains(body["error"], expected) {
		t.Fatalf("expected error containing %q, got %q", expected, body["error"])
	}
}

func AssertNonEmptyError(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()

	body := DecodeJSON[map[string]string](t, resp)
	if body["error"] == "" {
		t.Fatalf("expected non-empty error, got %+v", body)
	}
}
