package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/application/dto"
	"github.com/gachal/mossbase/services/rag/internal/application/service"
	"github.com/gachal/mossbase/services/rag/pkg/response"
)

// SearchHandler handles HTTP requests for similarity search.
type SearchHandler struct {
	docSvc service.DocumentService
}

// NewSearchHandler creates a new SearchHandler.
func NewSearchHandler(docSvc service.DocumentService) *SearchHandler {
	return &SearchHandler{
		docSvc: docSvc,
	}
}

// Search handles POST /api/v1/search.
// It binds the JSON request, validates the query, and performs similarity search.
func (h *SearchHandler) Search(c *gin.Context) {
	var req dto.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.Query == "" {
		response.Error(c, http.StatusBadRequest, "query is required")
		return
	}
	if req.SpaceID == "" {
		response.Error(c, http.StatusBadRequest, "space_id is required")
		return
	}

	// Set default TopK if not provided
	if req.TopK <= 0 {
		req.TopK = 10
	}

	result, err := h.docSvc.Search(c.Request.Context(), req)
	if err != nil {
		zap.L().Error("failed to search",
			zap.String("space_id", req.SpaceID),
			zap.String("query", req.Query),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, "failed to search")
		return
	}

	response.Success(c, result)
}
