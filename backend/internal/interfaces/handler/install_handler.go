package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/backend/internal/application/dto"
	"github.com/gachal/mossbase/backend/internal/application/service"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type InstallHandler struct {
	installService *service.InstallService
	shutdownCh     chan struct{}
}

func NewInstallHandler(installService *service.InstallService, shutdownCh chan struct{}) *InstallHandler {
	return &InstallHandler{
		installService: installService,
		shutdownCh:     shutdownCh,
	}
}

func (h *InstallHandler) GetStatus(c *gin.Context) {
	response.Success(c, dto.InstallStatusResponse{
		Installed: h.installService.GetStatus(),
	})
}

func (h *InstallHandler) TestDatabase(c *gin.Context) {
	var req dto.DatabaseInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.installService.TestDatabase(c.Request.Context(), req)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, result)
}

func (h *InstallHandler) Execute(c *gin.Context) {
	if h.installService.GetStatus() {
		response.Error(c, http.StatusNotFound, "not found")
		return
	}

	var req dto.InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.installService.Execute(c.Request.Context(), req); err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"installed": true,
		"message":   "Installation completed. Server is restarting...",
	})

	if h.shutdownCh != nil {
		select {
		case h.shutdownCh <- struct{}{}:
		default:
		}
	}
}
