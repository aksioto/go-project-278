package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"code/internal/domain/link"
	"code/internal/transport/http/validation"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestErrorsMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setupHandler   func(c *gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "no errors",
			setupHandler: func(c *gin.Context) {
				c.Status(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name: "validation error",
			setupHandler: func(c *gin.Context) {
				err := &validation.ValidationError{
					Fields: map[string]string{
						"original_url": "required",
					},
				}
				_ = c.Error(err)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"errors":{"original_url":"required"}}`,
		},
		{
			name: "invalid id error",
			setupHandler: func(c *gin.Context) {
				_ = c.Error(link.ErrInvalidID)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid id"}`,
		},
		{
			name: "link not found error",
			setupHandler: func(c *gin.Context) {
				_ = c.Error(link.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"link not found"}`,
		},
		{
			name: "visit not found error",
			setupHandler: func(c *gin.Context) {
				_ = c.Error(link.ErrVisitNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `{"error":"visit not found"}`,
		},
		{
			name: "short name taken error",
			setupHandler: func(c *gin.Context) {
				_ = c.Error(link.ErrShortNameTaken)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"errors":{"short_name":"short name already in use"}}`,
		},
		{
			name: "invalid original url error",
			setupHandler: func(c *gin.Context) {
				_ = c.Error(link.ErrInvalidOriginalURL)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedBody:   `{"errors":{"original_url":"cannot shorten own short URLs"}}`,
		},
		{
			name: "internal server error",
			setupHandler: func(c *gin.Context) {
				_ = c.Error(errors.New("unexpected error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorsMiddleware())
			router.GET("/test", tt.setupHandler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
			if tt.expectedBody != "" {
				assert.JSONEq(t, tt.expectedBody, resp.Body.String())
			}
		})
	}
}

func TestErrorsMiddleware_WrappedErrors(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "wrapped not found",
			err:            errors.Join(errors.New("context"), link.ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "wrapped invalid id",
			err:            errors.Join(errors.New("parse error"), link.ErrInvalidID),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorsMiddleware())
			router.GET("/test", func(c *gin.Context) {
				_ = c.Error(tt.err)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			resp := httptest.NewRecorder()

			router.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedStatus, resp.Code)
		})
	}
}
