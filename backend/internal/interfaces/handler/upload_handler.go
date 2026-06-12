package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gachal/mossbase/backend/internal/infrastructure/config"
	"github.com/gachal/mossbase/backend/pkg/upload"
	"github.com/gachal/mossbase/backend/pkg/response"
)

type UploadHandler struct {
	uploadDir   string
	maxFileSize int64
	baseURL     string
}

func NewUploadHandler(cfg *config.UploadConfig) *UploadHandler {
	maxSize := int64(5 * 1024 * 1024)
	if cfg.MaxSizeMB > 0 {
		maxSize = int64(cfg.MaxSizeMB) * 1024 * 1024
	}
	return &UploadHandler{
		uploadDir:   cfg.Dir,
		maxFileSize: maxSize,
		baseURL:     cfg.BaseURL,
	}
}

func (h *UploadHandler) UploadAvatar(c *gin.Context) {
	h.handleUpload(c, "avatars")
}

func (h *UploadHandler) UploadSpaceCover(c *gin.Context) {
	h.handleUpload(c, "spaces")
}

func (h *UploadHandler) handleUpload(c *gin.Context, subDir string) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, "请选择要上传的文件")
		return
	}
	defer file.Close()

	relPath, err := upload.SaveImage(file, header.Size, header.Filename, h.uploadDir, subDir, h.maxFileSize)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, gin.H{"url": h.baseURL + "/" + relPath})
}
