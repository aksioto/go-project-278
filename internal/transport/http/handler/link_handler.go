package handler

import (
	"errors"
	"net/http"
	"strconv"

	"code/internal/domain/link"
	"code/internal/transport/http/dto"
	"code/internal/transport/http/mapper"
	linkusecase "code/internal/usecase/link"

	"github.com/gin-gonic/gin"
)

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

func (h *LinkHandler) ListLinks(c *gin.Context) {
	links, err := h.service.ListLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := mapper.ToLinkResponseList(links, h.baseURL)
	c.JSON(http.StatusOK, response)
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
