package testutil

import (
	"net/http/httptest"
	"testing"
)

func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expected string) {
	t.Helper()

	body := DecodeJSON[map[string]string](t, resp)
	if body["error"] != expected {
		t.Fatalf("expected error %q, got %q", expected, body["error"])
	}
}

func AssertNonEmptyError(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()

	body := DecodeJSON[map[string]string](t, resp)
	if body["error"] == "" {
		t.Fatalf("expected non-empty error, got %+v", body)
	}
}
