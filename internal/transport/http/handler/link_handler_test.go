package handler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"code/internal/domain/link"
	"code/internal/testutil"
	"code/internal/transport/http/dto"
	"code/internal/transport/http/middleware"
	"code/internal/transport/http/validation"
	linkusecase "code/internal/usecase/link"
	visitusecase "code/internal/usecase/visit"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var testTimestamp = time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)

type createLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name,omitempty"`
}

type updateLinkRequest struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

func makeLink(id int64, originalURL, shortName string) link.Link {
	return link.Link{
		ID:          id,
		OriginalURL: originalURL,
		ShortName:   shortName,
		CreatedAt:   testTimestamp,
	}
}

func TestDeleteLinkVisit(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		visitMockSetup func(*visitusecase.MockService)
		expectedStatus int
	}{
		{
			name: "success",
			id:   "1",
			visitMockSetup: func(s *visitusecase.MockService) {
				s.EXPECT().DeleteVisit(gomock.Any(), int64(1)).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "service error",
			id:   "2",
			visitMockSetup: func(s *visitusecase.MockService) {
				s.EXPECT().DeleteVisit(gomock.Any(), int64(2)).Return(link.ErrVisitNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid id",
			id:             "abc",
			visitMockSetup: nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, nil, tt.visitMockSetup)

			resp := testutil.PerformRequest(t, router, http.MethodDelete, "/api/link_visits/"+tt.id, nil)

			require.Equal(t, tt.expectedStatus, resp.Code)
		})
	}
}

func linkToDTO(l link.Link) dto.LinkResponse {
	return dto.LinkResponse{
		ID:          l.ID,
		OriginalURL: l.OriginalURL,
		ShortName:   l.ShortName,
		ShortURL:    fmt.Sprintf("%s/r/%s", testutil.TestBaseURL, l.ShortName),
		CreatedAt:   l.CreatedAt,
	}
}

func makeLinkDTO(id int64, url, short string) dto.LinkResponse {
	return linkToDTO(makeLink(id, url, short))
}

func makeVisit(id, linkID int64, ip, userAgent, referer string, status int) link.Visit {
	return link.Visit{
		ID:        id,
		LinkID:    linkID,
		IP:        ip,
		UserAgent: userAgent,
		Referer:   referer,
		Status:    status,
		CreatedAt: testTimestamp,
	}
}

func visitToDTO(v link.Visit) dto.VisitResponse {
	return dto.VisitResponse{
		ID:        v.ID,
		LinkID:    v.LinkID,
		IP:        v.IP,
		UserAgent: v.UserAgent,
		Referer:   v.Referer,
		Status:    v.Status,
		CreatedAt: v.CreatedAt,
	}
}

func setupTestRouter(
	t *testing.T,
	linkMockExpect func(*linkusecase.MockService),
	visitMockExpect func(*visitusecase.MockService),
) *gin.Engine {
	t.Helper()

	ctrl := testutil.NewController(t)
	linkService := linkusecase.NewMockService(ctrl)
	visitService := visitusecase.NewMockService(ctrl)

	if linkMockExpect != nil {
		linkMockExpect(linkService)
	}
	if visitMockExpect != nil {
		visitMockExpect(visitService)
	}

	validation.Init()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}))

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare
	router.Use(middleware.ErrorsMiddleware())

	linkHandler := NewLinkHandler(linkService, visitService, testutil.TestBaseURL, logger)
	api := router.Group("/api")
	links := api.Group("/links")

	links.GET("", linkHandler.ListLinks)
	links.POST("", linkHandler.CreateLink)
	links.GET("/:id", linkHandler.GetLink)
	links.PUT("/:id", linkHandler.UpdateLink)
	links.DELETE("/:id", linkHandler.DeleteLink)

	visits := api.Group("/link_visits")
	visits.GET("", linkHandler.ListLinkVisits)
	visits.DELETE("/:id", linkHandler.DeleteLinkVisit)
	router.GET("/r/:code", linkHandler.RedirectToOriginalURL)

	return router
}

func TestCreateLink(t *testing.T) {
	tests := []struct {
		name           string
		body           *createLinkRequest
		mockExpect     func(*linkusecase.MockService)
		expectedStatus int
		assert         func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			body: &createLinkRequest{
				OriginalURL: "https://example.com",
				ShortName:   "exmpl",
			},
			mockExpect: func(s *linkusecase.MockService) {
				result := makeLink(1, "https://example.com", "exmpl")
				s.EXPECT().
					CreateLink(gomock.Any(), result.OriginalURL, result.ShortName).
					Return(&result, nil)
			},
			expectedStatus: http.StatusCreated,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := makeLinkDTO(1, "https://example.com", "exmpl")
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name: "validation error",
			body: &createLinkRequest{
				OriginalURL: "invalid",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertValidationErrors(t, resp, map[string]string{
					"original_url": "Key: 'CreateLinkRequest.original_url' Error:Field validation for 'original_url' failed on the 'rfc3986url' tag",
				})
			},
		},
		{
			name: "short name too short",
			body: &createLinkRequest{
				OriginalURL: "https://example.com",
				ShortName:   "ab",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertValidationErrors(t, resp, map[string]string{
					"short_name": "Key: 'CreateLinkRequest.short_name' Error:Field validation for 'short_name' failed on the 'min' tag",
				})
			},
		},
		{
			name: "conflict",
			body: &createLinkRequest{
				OriginalURL: "https://example.com",
				ShortName:   "exists",
			},
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					CreateLink(gomock.Any(), "https://example.com", "exists").
					Return(nil, link.ErrShortNameTaken)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertValidationErrors(t, resp, map[string]string{
					"short_name": "short name already in use",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect, nil)

			resp := testutil.PerformRequest(t, router, http.MethodPost, "/api/links", tt.body)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
		})
	}
}

func TestGetLink(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockExpect     func(*linkusecase.MockService)
		expectedStatus int
		assert         func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			id:   "1",
			mockExpect: func(s *linkusecase.MockService) {
				result := makeLink(1, "https://example.com", "ex")
				s.EXPECT().
					GetLink(gomock.Any(), int64(1)).
					Return(&result, nil)
			},
			expectedStatus: http.StatusOK,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := makeLinkDTO(1, "https://example.com", "ex")
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:           "invalid id",
			id:             "abc",
			mockExpect:     nil,
			expectedStatus: http.StatusBadRequest,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "invalid id")
			},
		},
		{
			name: "not found",
			id:   "10",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					GetLink(gomock.Any(), int64(10)).
					Return(nil, link.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "link not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect, nil)

			resp := testutil.PerformRequest(t, router, http.MethodGet, "/api/links/"+tt.id, nil)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
		})
	}
}

func TestListLinks(t *testing.T) {
	tests := []struct {
		name                 string
		rangeParam           string
		mockExpect           func(*linkusecase.MockService)
		expectedStatus       int
		expectedContentRange string
		assert               func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:       "success without range (default pagination)",
			rangeParam: "",
			mockExpect: func(s *linkusecase.MockService) {
				results := []link.Link{
					makeLink(1, "https://example.com/1", "one"),
					makeLink(2, "https://example.com/2", "two"),
				}
				s.EXPECT().
					ListLinksPaginated(gomock.Any(), defaultPageSize, int32(0)).
					Return(&linkusecase.ListResult{Links: results, Total: 2}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "links 0-1/2",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.LinkResponse{
					makeLinkDTO(1, "https://example.com/1", "one"),
					makeLinkDTO(2, "https://example.com/2", "two"),
				}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "with range parameter",
			rangeParam: "[0,9]",
			mockExpect: func(s *linkusecase.MockService) {
				results := []link.Link{
					makeLink(1, "https://example.com/1", "one"),
					makeLink(2, "https://example.com/2", "two"),
				}
				s.EXPECT().
					ListLinksPaginated(gomock.Any(), int32(10), int32(0)).
					Return(&linkusecase.ListResult{Links: results, Total: 42}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "links 0-1/42",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.LinkResponse{
					makeLinkDTO(1, "https://example.com/1", "one"),
					makeLinkDTO(2, "https://example.com/2", "two"),
				}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "second page",
			rangeParam: "[5,6]",
			mockExpect: func(s *linkusecase.MockService) {
				results := []link.Link{
					makeLink(6, "https://example.com/6", "six"),
					makeLink(7, "https://example.com/7", "seven"),
				}
				s.EXPECT().
					ListLinksPaginated(gomock.Any(), int32(2), int32(5)).
					Return(&linkusecase.ListResult{Links: results, Total: 11}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "links 5-6/11",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.LinkResponse{
					makeLinkDTO(6, "https://example.com/6", "six"),
					makeLinkDTO(7, "https://example.com/7", "seven"),
				}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "empty result",
			rangeParam: "[100,109]",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					ListLinksPaginated(gomock.Any(), int32(10), int32(100)).
					Return(&linkusecase.ListResult{Links: []link.Link{}, Total: 42}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "links 100-100/42",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.LinkResponse{}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "invalid range format returns 416",
			rangeParam: "invalid",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					CountLinks(gomock.Any()).
					Return(int64(1), nil)
			},
			expectedStatus:       http.StatusRequestedRangeNotSatisfiable,
			expectedContentRange: "links */1",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "invalid range")
			},
		},
		{
			name:       "negative or reversed range returns 416",
			rangeParam: "[5,-1]",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					CountLinks(gomock.Any()).
					Return(int64(42), nil)
			},
			expectedStatus:       http.StatusRequestedRangeNotSatisfiable,
			expectedContentRange: "links */42",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "invalid range")
			},
		},
		{
			name:       "invalid range when count fails",
			rangeParam: "[bad]",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					CountLinks(gomock.Any()).
					Return(int64(0), fmt.Errorf("count failure"))
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "internal server error")
			},
		},
		{
			name:       "range larger than max page size clamps limit",
			rangeParam: "[0,500]",
			mockExpect: func(s *linkusecase.MockService) {
				results := []link.Link{
					makeLink(1, "https://example.com/1", "one"),
					makeLink(2, "https://example.com/2", "two"),
				}
				s.EXPECT().
					ListLinksPaginated(gomock.Any(), maxPageSize, int32(0)).
					Return(&linkusecase.ListResult{Links: results, Total: 500}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "links 0-1/500",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.LinkResponse{
					makeLinkDTO(1, "https://example.com/1", "one"),
					makeLinkDTO(2, "https://example.com/2", "two"),
				}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "service error",
			rangeParam: "",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					ListLinksPaginated(gomock.Any(), defaultPageSize, int32(0)).
					Return(nil, errors.New("db failure"))
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "internal server error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect, nil)

			path := "/api/links"
			if tt.rangeParam != "" {
				path += "?range=" + tt.rangeParam
			}
			resp := testutil.PerformRequest(t, router, http.MethodGet, path, nil)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
			if tt.expectedContentRange != "" {
				assert.Equal(t, tt.expectedContentRange, resp.Header().Get("Content-Range"))
			}
		})
	}
}

func TestUpdateLink(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		body           *updateLinkRequest
		mockExpect     func(*linkusecase.MockService)
		expectedStatus int
		assert         func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			id:   "1",
			body: &updateLinkRequest{
				OriginalURL: "https://example.com/updated",
				ShortName:   "upd",
			},
			mockExpect: func(s *linkusecase.MockService) {
				result := makeLink(1, "https://example.com/updated", "upd")
				s.EXPECT().
					UpdateLink(gomock.Any(), int64(1), result.OriginalURL, result.ShortName).
					Return(&result, nil)
			},
			expectedStatus: http.StatusOK,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := makeLinkDTO(1, "https://example.com/updated", "upd")
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:           "invalid id",
			id:             "abc",
			body:           &updateLinkRequest{},
			mockExpect:     nil,
			expectedStatus: http.StatusBadRequest,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "invalid id")
			},
		},
		{
			name: "validation error",
			id:   "1",
			body: &updateLinkRequest{
				OriginalURL: "bad",
				ShortName:   "upd",
			},
			mockExpect:     nil,
			expectedStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertValidationErrors(t, resp, map[string]string{
					"original_url": "Key: 'UpdateLinkRequest.original_url' Error:Field validation for 'original_url' failed on the 'rfc3986url' tag",
				})
			},
		},
		{
			name: "not found",
			id:   "1",
			body: &updateLinkRequest{
				OriginalURL: "https://example.com/updated",
				ShortName:   "upd",
			},
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					UpdateLink(gomock.Any(), int64(1), "https://example.com/updated", "upd").
					Return(nil, link.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "link not found")
			},
		},
		{
			name: "conflict",
			id:   "1",
			body: &updateLinkRequest{
				OriginalURL: "https://example.com/updated",
				ShortName:   "dup",
			},
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					UpdateLink(gomock.Any(), int64(1), "https://example.com/updated", "dup").
					Return(nil, link.ErrShortNameTaken)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertValidationErrors(t, resp, map[string]string{
					"short_name": "short name already in use",
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect, nil)

			resp := testutil.PerformRequest(t, router, http.MethodPut, "/api/links/"+tt.id, tt.body)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
		})
	}
}

func TestDeleteLink(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		mockExpect     func(*linkusecase.MockService)
		expectedStatus int
		assert         func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			id:   "1",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					DeleteLink(gomock.Any(), int64(1)).
					Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id",
			id:             "abc",
			mockExpect:     nil,
			expectedStatus: http.StatusBadRequest,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "invalid id")
			},
		},
		{
			name: "not found",
			id:   "1",
			mockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					DeleteLink(gomock.Any(), int64(1)).
					Return(link.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect, nil)

			resp := testutil.PerformRequest(t, router, http.MethodDelete, "/api/links/"+tt.id, nil)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
		})
	}
}

func TestListLinkVisits(t *testing.T) {
	tests := []struct {
		name                 string
		rangeParam           string
		visitMockExpect      func(*visitusecase.MockService)
		expectedStatus       int
		expectedContentRange string
		assert               func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name:       "success default pagination",
			rangeParam: "",
			visitMockExpect: func(s *visitusecase.MockService) {
				visits := []link.Visit{
					makeVisit(1, 10, "1.1.1.1", "ua", "ref", http.StatusFound),
					makeVisit(2, 11, "2.2.2.2", "ua2", "", http.StatusFound),
				}
				s.EXPECT().
					ListVisitsPaginated(gomock.Any(), defaultPageSize, int32(0)).
					Return(&visitusecase.ListResult{Visits: visits, Total: 2}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "link_visits 0-1/2",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.VisitResponse{
					visitToDTO(makeVisit(1, 10, "1.1.1.1", "ua", "ref", http.StatusFound)),
					visitToDTO(makeVisit(2, 11, "2.2.2.2", "ua2", "", http.StatusFound)),
				}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "with range parameter",
			rangeParam: "[5,6]",
			visitMockExpect: func(s *visitusecase.MockService) {
				visits := []link.Visit{
					makeVisit(3, 12, "3.3.3.3", "ua3", "", http.StatusFound),
				}
				s.EXPECT().
					ListVisitsPaginated(gomock.Any(), int32(2), int32(5)).
					Return(&visitusecase.ListResult{Visits: visits, Total: 7}, nil)
			},
			expectedStatus:       http.StatusOK,
			expectedContentRange: "link_visits 5-5/7",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				expected := []dto.VisitResponse{
					visitToDTO(makeVisit(3, 12, "3.3.3.3", "ua3", "", http.StatusFound)),
				}
				testutil.AssertJSON(t, resp, expected)
			},
		},
		{
			name:       "invalid range format",
			rangeParam: "bad",
			visitMockExpect: func(s *visitusecase.MockService) {
				s.EXPECT().
					CountVisits(gomock.Any()).
					Return(int64(4), nil)
			},
			expectedStatus:       http.StatusRequestedRangeNotSatisfiable,
			expectedContentRange: "link_visits */4",
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "invalid range")
			},
		},
		{
			name:       "count error on invalid range",
			rangeParam: "oops",
			visitMockExpect: func(s *visitusecase.MockService) {
				s.EXPECT().
					CountVisits(gomock.Any()).
					Return(int64(0), fmt.Errorf("count failure"))
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "internal server error")
			},
		},
		{
			name:       "service error",
			rangeParam: "",
			visitMockExpect: func(s *visitusecase.MockService) {
				s.EXPECT().
					ListVisitsPaginated(gomock.Any(), defaultPageSize, int32(0)).
					Return(nil, fmt.Errorf("db failure"))
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "internal server error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, nil, tt.visitMockExpect)

			path := "/api/link_visits"
			if tt.rangeParam != "" {
				path += "?range=" + tt.rangeParam
			}

			resp := testutil.PerformRequest(t, router, http.MethodGet, path, nil)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
			assert.Equal(t, tt.expectedContentRange, resp.Header().Get("Content-Range"))
		})
	}
}

func TestRedirectToOriginalURL(t *testing.T) {
	const (
		code            = "abc123"
		expectedIP      = "203.0.113.1"
		expectedUA      = "test-agent"
		expectedReferer = "https://ref.example"
	)

	tests := []struct {
		name            string
		linkMockExpect  func(*linkusecase.MockService)
		visitMockExpect func(*visitusecase.MockService)
		expectedStatus  int
		assert          func(t *testing.T, resp *httptest.ResponseRecorder)
	}{
		{
			name: "success",
			linkMockExpect: func(s *linkusecase.MockService) {
				result := makeLink(1, "https://example.com", code)
				s.EXPECT().
					GetLinkByShortName(gomock.Any(), code).
					Return(&result, nil)
			},
			visitMockExpect: func(s *visitusecase.MockService) {
				s.EXPECT().
					CreateVisit(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, visit link.Visit) (*link.Visit, error) {
						if visit.LinkID != 1 {
							return nil, fmt.Errorf("unexpected link id %d", visit.LinkID)
						}
						if visit.Status != http.StatusFound {
							return nil, fmt.Errorf("unexpected status %d", visit.Status)
						}
						return &visit, nil
					})
			},
			expectedStatus: http.StatusFound,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, "https://example.com", resp.Header().Get("Location"))
			},
		},
		{
			name: "link not found",
			linkMockExpect: func(s *linkusecase.MockService) {
				s.EXPECT().
					GetLinkByShortName(gomock.Any(), code).
					Return(nil, link.ErrNotFound)
			},
			visitMockExpect: nil,
			expectedStatus:  http.StatusNotFound,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "link not found")
			},
		},
		{
			name: "visit creation error - redirect still works",
			linkMockExpect: func(s *linkusecase.MockService) {
				result := makeLink(1, "https://example.com", code)
				s.EXPECT().
					GetLinkByShortName(gomock.Any(), code).
					Return(&result, nil)
			},
			visitMockExpect: func(s *visitusecase.MockService) {
				s.EXPECT().
					CreateVisit(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("save failure"))
			},
			expectedStatus: http.StatusFound,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				assert.Equal(t, "https://example.com", resp.Header().Get("Location"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.linkMockExpect, tt.visitMockExpect)

			headers := map[string]string{
				"Cf-Connecting-Ip": expectedIP,
				"User-Agent":       expectedUA,
				"Referer":          expectedReferer,
			}

			resp := testutil.PerformRequestWithHeaders(t, router, http.MethodGet, "/r/"+code, nil, headers)
			require.Equal(t, tt.expectedStatus, resp.Code)
			if tt.assert != nil {
				tt.assert(t, resp)
			}
		})
	}
}
