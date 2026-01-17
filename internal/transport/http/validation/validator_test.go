package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRFC3986URL(t *testing.T) {
	v := validator.New()
	err := RegisterCustomValidators(v)
	require.NoError(t, err)

	tests := []struct {
		name    string
		url     string
		isValid bool
	}{
		{
			name:    "valid http url",
			url:     "http://example.com",
			isValid: true,
		},
		{
			name:    "valid https url",
			url:     "https://example.com",
			isValid: true,
		},
		{
			name:    "valid url with path",
			url:     "https://example.com/path/to/resource",
			isValid: true,
		},
		{
			name:    "valid url with query",
			url:     "https://example.com/search?q=test&page=1",
			isValid: true,
		},
		{
			name:    "valid url with port",
			url:     "https://example.com:8080/api",
			isValid: true,
		},
		{
			name:    "valid url with fragment",
			url:     "https://example.com/page#section",
			isValid: true,
		},
		{
			name:    "valid localhost",
			url:     "http://localhost:3000",
			isValid: true,
		},
		{
			name:    "valid ip address",
			url:     "http://192.168.1.1:8080",
			isValid: true,
		},
		{
			name:    "invalid - empty string",
			url:     "",
			isValid: false,
		},
		{
			name:    "invalid - no scheme",
			url:     "example.com",
			isValid: false,
		},
		{
			name:    "invalid - ftp scheme",
			url:     "ftp://example.com",
			isValid: false,
		},
		{
			name:    "invalid - file scheme",
			url:     "file:///path/to/file",
			isValid: false,
		},
		{
			name:    "invalid - javascript scheme",
			url:     "javascript:alert(1)",
			isValid: false,
		},
		{
			name:    "invalid - missing host",
			url:     "https://",
			isValid: false,
		},
		{
			name:    "invalid - relative path",
			url:     "/path/to/resource",
			isValid: false,
		},
		{
			name:    "invalid - malformed url",
			url:     "not a url at all",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			type testStruct struct {
				URL string `validate:"rfc3986url"`
			}

			s := testStruct{URL: tt.url}
			err := v.Struct(s)

			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
