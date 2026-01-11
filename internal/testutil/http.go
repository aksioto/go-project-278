package testutil

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func PerformRequest(t *testing.T, handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &buf)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	return w
}

func AssertStatus(t *testing.T, got, want int) {
	t.Helper()

	if got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
}

func AssertResponse(t *testing.T, resp *httptest.ResponseRecorder, status int, assert func(t *testing.T, resp *httptest.ResponseRecorder)) {
	t.Helper()

	AssertStatus(t, resp.Code, status)

	if assert != nil {
		assert(t, resp)
	}
}
