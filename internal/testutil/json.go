package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func DecodeJSON[T any](t *testing.T, resp *httptest.ResponseRecorder) T {
	t.Helper()

	var dst T
	if err := json.NewDecoder(resp.Body).Decode(&dst); err != nil {
		t.Fatalf("decode json: %v (body=%s)", err, resp.Body.String())
	}
	return dst
}

func AssertJSON[T any](t *testing.T, resp *httptest.ResponseRecorder, want T) {
	t.Helper()

	got := DecodeJSON[T](t, resp)
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("response mismatch (-want +got):\n%s", diff)
	}
}
