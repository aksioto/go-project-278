package handler

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"code/internal/domain/link"
	"code/internal/testutil"
	"code/internal/transport/http/dto"
	linkusecase "code/internal/usecase/link"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
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

func setupTestRouter(t *testing.T, mockExpect func(*linkusecase.MockService)) *gin.Engine {
	t.Helper()

	ctrl := testutil.NewController(t)
	service := linkusecase.NewMockService(ctrl)

	if mockExpect != nil {
		mockExpect(service)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.TrustedPlatform = gin.PlatformCloudflare

	linkHandler := NewLinkHandler(service, testutil.TestBaseURL)
	api := router.Group("/api")
	links := api.Group("/links")

	links.GET("", linkHandler.ListLinks)
	links.POST("", linkHandler.CreateLink)
	links.GET("/:id", linkHandler.GetLink)
	links.PUT("/:id", linkHandler.UpdateLink)
	links.DELETE("/:id", linkHandler.DeleteLink)

	api.GET("/link_visits", linkHandler.ListLinkVisits)
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
			expectedStatus: http.StatusBadRequest,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "Key: 'CreateLinkRequest.OriginalURL' Error:Field validation for 'OriginalURL' failed on the 'url' tag")
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
			expectedStatus: http.StatusConflict,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "short_name already exists")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect)

			resp := testutil.PerformRequest(t, router, http.MethodPost, "/api/links", tt.body)
			testutil.AssertResponse(t, resp, tt.expectedStatus, tt.assert)
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
			router := setupTestRouter(t, tt.mockExpect)

			resp := testutil.PerformRequest(t, router, http.MethodGet, "/api/links/"+tt.id, nil)
			testutil.AssertResponse(t, resp, tt.expectedStatus, tt.assert)
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
			expectedStatus:       http.StatusPartialContent,
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
			expectedStatus:       http.StatusPartialContent,
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
			expectedStatus:       http.StatusPartialContent,
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
					Return(int64(0), errors.New("count failure"))
			},
			expectedStatus: http.StatusInternalServerError,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "count failure")
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
			expectedStatus:       http.StatusPartialContent,
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
				testutil.AssertErrorResponse(t, resp, "db failure")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect)

			path := "/api/links"
			if tt.rangeParam != "" {
				path += "?range=" + tt.rangeParam
			}
			resp := testutil.PerformRequest(t, router, http.MethodGet, path, nil)
			testutil.AssertResponse(t, resp, tt.expectedStatus, tt.assert)

			if tt.expectedContentRange != "" {
				got := resp.Header().Get("Content-Range")
				if got != tt.expectedContentRange {
					t.Fatalf("expected Content-Range %q, got %q", tt.expectedContentRange, got)
				}
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
			expectedStatus: http.StatusBadRequest,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "Key: 'UpdateLinkRequest.OriginalURL' Error:Field validation for 'OriginalURL' failed on the 'url' tag")
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
			expectedStatus: http.StatusConflict,
			assert: func(t *testing.T, resp *httptest.ResponseRecorder) {
				testutil.AssertErrorResponse(t, resp, "short_name already exists")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupTestRouter(t, tt.mockExpect)

			resp := testutil.PerformRequest(t, router, http.MethodPut, "/api/links/"+tt.id, tt.body)
			testutil.AssertResponse(t, resp, tt.expectedStatus, tt.assert)
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
			router := setupTestRouter(t, tt.mockExpect)

			resp := testutil.PerformRequest(t, router, http.MethodDelete, "/api/links/"+tt.id, nil)
			testutil.AssertResponse(t, resp, tt.expectedStatus, tt.assert)
		})
	}
}
