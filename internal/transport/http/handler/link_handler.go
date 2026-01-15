package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"code/internal/domain/link"
	"code/internal/transport/http/dto"
	"code/internal/transport/http/mapper"
	linkusecase "code/internal/usecase/link"

	"github.com/gin-gonic/gin"
)

const (
	defaultPageSize int32 = 50
	maxPageSize     int32 = 200
)

type paginationRange struct {
	start int32
	end   int32
}

type LinkHandler struct {
	service linkusecase.Service
	baseURL string
}

func NewLinkHandler(service linkusecase.Service, baseURL string) *LinkHandler {
	return &LinkHandler{
		service: service,
		baseURL: baseURL,
	}
}

func (h *LinkHandler) parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func (h *LinkHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, link.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "link not found"})
	case errors.Is(err, link.ErrShortNameTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "short_name already exists"})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}

func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req dto.CreateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domainLink, err := h.service.CreateLink(c.Request.Context(), req.OriginalURL, req.ShortName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, mapper.ToLinkResponse(*domainLink, h.baseURL))
}

func (h *LinkHandler) GetLink(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	domainLink, err := h.service.GetLink(c.Request.Context(), id)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.ToLinkResponse(*domainLink, h.baseURL))
}

func (h *LinkHandler) parseRange(rangeParam string) (paginationRange, error) {
	if rangeParam == "" {
		return paginationRange{start: 0, end: defaultPageSize - 1}, nil
	}

	var rangeValues []int32
	if err := json.Unmarshal([]byte(rangeParam), &rangeValues); err != nil || len(rangeValues) != 2 {
		return paginationRange{}, fmt.Errorf("invalid range format: %s", rangeParam)
	}

	start, end := rangeValues[0], rangeValues[1]
	if start < 0 || end < 0 || start > end {
		return paginationRange{}, fmt.Errorf("invalid range: %s", rangeParam)
	}

	if span := end - start + 1; span > maxPageSize {
		end = start + maxPageSize - 1
	}

	return paginationRange{start: start, end: end}, nil
}

func (h *LinkHandler) ListLinks(c *gin.Context) {
	rangeParam := c.Query("range")
	pRange, err := h.parseRange(rangeParam)
	if err != nil {
		total, countErr := h.service.CountLinks(c.Request.Context())
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": countErr.Error()})
			return
		}
		c.Header("Content-Range", fmt.Sprintf("links */%d", total))
		c.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": err.Error()})
		return
	}

	limit := pRange.end - pRange.start + 1
	offset := pRange.start

	result, err := h.service.ListLinksPaginated(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	actualEnd := offset + int32(len(result.Links)) - 1
	if len(result.Links) == 0 {
		actualEnd = offset
	}
	c.Header("Content-Range", fmt.Sprintf("links %d-%d/%d", pRange.start, actualEnd, result.Total))

	response := mapper.ToLinkResponseList(result.Links, h.baseURL)
	status := http.StatusOK
	if rangeParam != "" {
		status = http.StatusPartialContent
	}

	c.JSON(status, response)
}

func (h *LinkHandler) ListLinkVisits(c *gin.Context) {
	rangeParam := c.Query("range")
	pRange, err := h.parseRange(rangeParam)
	if err != nil {
		total, countErr := h.service.CountLinkVisits(c.Request.Context())
		if countErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": countErr.Error()})
			return
		}
		c.Header("Content-Range", fmt.Sprintf("link_visits */%d", total))
		c.JSON(http.StatusRequestedRangeNotSatisfiable, gin.H{"error": err.Error()})
		return
	}

	limit := pRange.end - pRange.start + 1
	offset := pRange.start

	result, err := h.service.ListLinkVisitsPaginated(c.Request.Context(), limit, offset)
	if err != nil {
		h.handleError(c, err)
		return
	}

	actualEnd := offset + int32(len(result.Visits)) - 1
	if len(result.Visits) == 0 {
		actualEnd = offset
	}
	c.Header("Content-Range", fmt.Sprintf("link_visits %d-%d/%d", pRange.start, actualEnd, result.Total))

	response := mapper.ToVisitResponseList(result.Visits)
	status := http.StatusOK
	if rangeParam != "" {
		status = http.StatusPartialContent
	}

	c.JSON(status, response)
}

func (h *LinkHandler) UpdateLink(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domainLink, err := h.service.UpdateLink(c.Request.Context(), id, req.OriginalURL, req.ShortName)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, mapper.ToLinkResponse(*domainLink, h.baseURL))
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	id, ok := h.parseID(c)
	if !ok {
		return
	}

	if err := h.service.DeleteLink(c.Request.Context(), id); err != nil {
		h.handleError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *LinkHandler) RedirectToOriginalURL(c *gin.Context) {
	code := c.Param("code")

	domainLink, err := h.service.GetLinkByShortName(c.Request.Context(), code)
	if err != nil {
		h.handleError(c, err)
		return
	}

	status := http.StatusFound
	visit := link.Visit{
		LinkID:    domainLink.ID,
		IP:        c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Referer:   c.Request.Referer(),
		Status:    status,
	}

	if _, err := h.service.CreateLinkVisit(c.Request.Context(), visit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(status, domainLink.OriginalURL)
}
