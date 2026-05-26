package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type PageHandler struct {
	pageService service.PageService
}

func NewPageHandler(pageService service.PageService) *PageHandler {
	return &PageHandler{pageService: pageService}
}

func (h *PageHandler) Create(c *gin.Context) {
	spaceID, err := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid space id")
		return
	}

	var req dto.CreatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := response.GetUserID(c)
	result, err := h.pageService.Create(c.Request.Context(), spaceID, userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, result)
}

func (h *PageHandler) Update(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}

	var req dto.UpdatePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := response.GetUserID(c)
	result, err := h.pageService.Update(c.Request.Context(), pageID, userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PageHandler) Delete(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}

	if err := h.pageService.Delete(c.Request.Context(), pageID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *PageHandler) GetByID(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}

	result, err := h.pageService.GetByID(c.Request.Context(), pageID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "page not found")
		return
	}
	response.Success(c, result)
}

func (h *PageHandler) GetTree(c *gin.Context) {
	spaceID, err := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid space id")
		return
	}

	result, err := h.pageService.GetTreeBySpace(c.Request.Context(), spaceID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PageHandler) Search(c *gin.Context) {
	spaceID, err := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid space id")
		return
	}

	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "search query is required")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	result, err := h.pageService.Search(c.Request.Context(), spaceID, query, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PageHandler) MovePage(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}

	var req dto.MovePageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.pageService.MovePage(c.Request.Context(), pageID, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PageHandler) SemanticSearch(c *gin.Context) {
	spaceID, err := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid space id")
		return
	}

	query := c.Query("q")
	if query == "" {
		response.Error(c, http.StatusBadRequest, "query is required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	result, err := h.pageService.SemanticSearch(c.Request.Context(), spaceID, query, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}
