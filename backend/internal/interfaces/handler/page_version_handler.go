package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type PageVersionHandler struct {
	versionService service.PageVersionService
}

func NewPageVersionHandler(versionService service.PageVersionService) *PageVersionHandler {
	return &PageVersionHandler{versionService: versionService}
}

func (h *PageVersionHandler) ListVersions(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	versions, total, err := h.versionService.ListVersions(c.Request.Context(), pageID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, versions, total, page, pageSize)
}

func (h *PageVersionHandler) GetVersion(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}
	versionNum, err := strconv.Atoi(c.Param("versionNumber"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid version number")
		return
	}

	result, err := h.versionService.GetVersion(c.Request.Context(), pageID, versionNum)
	if err != nil {
		response.Error(c, http.StatusNotFound, "version not found")
		return
	}
	response.Success(c, result)
}

func (h *PageVersionHandler) GetDiff(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}
	from, _ := strconv.Atoi(c.Query("from"))
	to, _ := strconv.Atoi(c.Query("to"))
	if from == 0 || to == 0 {
		response.Error(c, http.StatusBadRequest, "from and to version numbers required")
		return
	}

	result, err := h.versionService.GetDiff(c.Request.Context(), pageID, from, to)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *PageVersionHandler) RestoreVersion(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}
	versionNum, err := strconv.Atoi(c.Param("versionNumber"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid version number")
		return
	}

	userID := response.GetUserID(c)
	result, err := h.versionService.RestoreVersion(c.Request.Context(), pageID, userID, versionNum)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

// Compile-time check
var _ dto.VersionDiffRequest
