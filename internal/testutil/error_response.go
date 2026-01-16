package testutil

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func AssertErrorResponse(t *testing.T, resp *httptest.ResponseRecorder, expected string) {
	t.Helper()

	body := DecodeJSON[map[string]string](t, resp)
	assert.Contains(t, body["error"], expected, "error response mismatch")
}

func AssertNonEmptyError(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()

	body := DecodeJSON[map[string]string](t, resp)
	require.NotEmpty(t, body["error"], "expected non-empty error")
}

func AssertValidationErrors(t *testing.T, resp *httptest.ResponseRecorder, expected map[string]string) {
	t.Helper()

	body := DecodeJSON[struct {
		Errors map[string]string `json:"errors"`
	}](t, resp)

	assert.Equal(t, expected, body.Errors, "validation errors mismatch")
}
