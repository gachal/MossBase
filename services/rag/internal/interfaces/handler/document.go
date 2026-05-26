package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/gachal/mossbase/services/rag/internal/application/dto"
	"github.com/gachal/mossbase/services/rag/internal/application/service"
	"github.com/gachal/mossbase/services/rag/pkg/response"
)

// DocumentHandler handles HTTP requests for document indexing and deletion.
type DocumentHandler struct {
	docSvc service.DocumentService
}

// NewDocumentHandler creates a new DocumentHandler.
func NewDocumentHandler(docSvc service.DocumentService) *DocumentHandler {
	return &DocumentHandler{
		docSvc: docSvc,
	}
}

// IndexDocument handles POST /api/v1/documents.
// It binds the JSON request, validates required fields, and indexes the document.
func (h *DocumentHandler) IndexDocument(c *gin.Context) {
	var req dto.IndexDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	if req.DocumentID == "" {
		response.Error(c, http.StatusBadRequest, "document_id is required")
		return
	}
	if req.SpaceID == "" {
		response.Error(c, http.StatusBadRequest, "space_id is required")
		return
	}
	if req.Title == "" {
		response.Error(c, http.StatusBadRequest, "title is required")
		return
	}
	if req.Content == "" {
		response.Error(c, http.StatusBadRequest, "content is required")
		return
	}

	if err := h.docSvc.IndexDocument(c.Request.Context(), req); err != nil {
		zap.L().Error("failed to index document",
			zap.String("document_id", req.DocumentID),
			zap.String("space_id", req.SpaceID),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, "failed to index document")
		return
	}

	response.Created(c, gin.H{
		"document_id": req.DocumentID,
		"space_id":    req.SpaceID,
	})
}

// DeleteDocument handles DELETE /api/v1/documents/:id.
// It parses the document ID from the path and space_id from the query string.
func (h *DocumentHandler) DeleteDocument(c *gin.Context) {
	documentID := c.Param("id")
	if documentID == "" {
		response.Error(c, http.StatusBadRequest, "document id is required")
		return
	}

	spaceID := c.Query("space_id")
	if spaceID == "" {
		response.Error(c, http.StatusBadRequest, "space_id query parameter is required")
		return
	}

	req := dto.DeleteDocumentRequest{
		DocumentID: documentID,
		SpaceID:    spaceID,
	}

	if err := h.docSvc.DeleteDocument(c.Request.Context(), req); err != nil {
		zap.L().Error("failed to delete document",
			zap.String("document_id", documentID),
			zap.String("space_id", spaceID),
			zap.Error(err),
		)
		response.Error(c, http.StatusInternalServerError, "failed to delete document")
		return
	}

	response.Success(c, gin.H{
		"document_id": documentID,
		"space_id":    spaceID,
	})
}
