package testutil

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func DecodeJSON[T any](t *testing.T, resp *httptest.ResponseRecorder) T {
	t.Helper()

	var dst T
	err := json.NewDecoder(resp.Body).Decode(&dst)
	require.NoError(t, err, "decode json, body=%s", resp.Body.String())
	return dst
}

func AssertJSON[T any](t *testing.T, resp *httptest.ResponseRecorder, want T) {
	t.Helper()

	got := DecodeJSON[T](t, resp)
	assert.Equal(t, want, got, "response JSON mismatch")
}
