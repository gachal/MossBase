package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	stats, err := h.adminService.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, stats)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	users, total, err := h.adminService.ListUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, users, total, page, pageSize)
}

func (h *AdminHandler) UpdateUserRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	var req dto.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.adminService.UpdateUserRole(c.Request.Context(), userID, req.Role); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}
	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.adminService.UpdateUserStatus(c.Request.Context(), userID, req.Status); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) ListSpaces(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	spaces, total, err := h.adminService.ListSpaces(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, spaces, total, page, pageSize)
}

func (h *AdminHandler) GetSpaceDetail(c *gin.Context) {
	spaceID, err := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid space id")
		return
	}
	detail, err := h.adminService.GetSpaceDetail(c.Request.Context(), spaceID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "space not found")
		return
	}
	response.Success(c, detail)
}

func (h *AdminHandler) DeleteSpace(c *gin.Context) {
	spaceID, err := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid space id")
		return
	}
	if err := h.adminService.DeleteSpace(c.Request.Context(), spaceID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) ListPages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	pages, total, err := h.adminService.ListPages(c.Request.Context(), page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, pages, total, page, pageSize)
}

func (h *AdminHandler) DeletePage(c *gin.Context) {
	pageID, err := strconv.ParseUint(c.Param("pageId"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid page id")
		return
	}
	if err := h.adminService.DeletePage(c.Request.Context(), pageID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) GetSettings(c *gin.Context) {
	settings, err := h.adminService.GetSettings(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "获取设置失败")
		return
	}
	response.Success(c, settings)
}

func (h *AdminHandler) UpdateSettings(c *gin.Context) {
	var req dto.SettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.adminService.UpdateSettings(c.Request.Context(), req); err != nil {
		response.Error(c, http.StatusInternalServerError, "保存设置失败")
		return
	}
	response.Success(c, nil)
}

func (h *AdminHandler) TestRAGConnection(c *gin.Context) {
	var req dto.TestRAGRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.adminService.TestRAGConnection(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "测试连接失败")
		return
	}
	response.Success(c, result)
}
