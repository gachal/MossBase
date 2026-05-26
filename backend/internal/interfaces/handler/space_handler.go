package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type SpaceHandler struct {
	spaceSvc service.SpaceService
}

func NewSpaceHandler(spaceSvc service.SpaceService) *SpaceHandler {
	return &SpaceHandler{spaceSvc: spaceSvc}
}

func (h *SpaceHandler) Create(c *gin.Context) {
	var req dto.CreateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	userID := response.GetUserID(c)
	result, err := h.spaceSvc.Create(c.Request.Context(), userID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, result)
}

func (h *SpaceHandler) List(c *gin.Context) {
	userID := response.GetUserID(c)
	page := c.GetInt("page")
	pageSize := c.GetInt("page_size")
	items, total, err := h.spaceSvc.ListByUser(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Paginated(c, items, total, page, pageSize)
}

func (h *SpaceHandler) GetByID(c *gin.Context) {
	spaceID, _ := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	result, err := h.spaceSvc.GetByID(c.Request.Context(), spaceID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "space not found")
		return
	}
	response.Success(c, result)
}

func (h *SpaceHandler) Update(c *gin.Context) {
	spaceID, _ := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	var req dto.UpdateSpaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	result, err := h.spaceSvc.Update(c.Request.Context(), spaceID, req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, result)
}

func (h *SpaceHandler) Delete(c *gin.Context) {
	spaceID, _ := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	if err := h.spaceSvc.Delete(c.Request.Context(), spaceID); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SpaceHandler) AddMember(c *gin.Context) {
	spaceID, _ := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	userID := response.GetUserID(c)
	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.spaceSvc.AddMember(c.Request.Context(), spaceID, userID, req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SpaceHandler) RemoveMember(c *gin.Context) {
	spaceID, _ := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	userID := response.GetUserID(c)
	targetID, _ := strconv.ParseUint(c.Param("userId"), 10, 64)
	if err := h.spaceSvc.RemoveMember(c.Request.Context(), spaceID, userID, targetID); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, nil)
}

func (h *SpaceHandler) ListMembers(c *gin.Context) {
	spaceID, _ := strconv.ParseUint(c.Param("spaceId"), 10, 64)
	members, err := h.spaceSvc.ListMembers(c.Request.Context(), spaceID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Success(c, members)
}
