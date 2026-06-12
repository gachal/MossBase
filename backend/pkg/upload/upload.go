package upload

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 不允许 SVG，因为 SVG 可包含 JavaScript 导致 XSS
var allowedMIMETypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

var mimeToExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

var allowedSubDirs = map[string]bool{
	"avatars": true,
	"spaces":  true,
}

func SaveImage(fileHeader io.Reader, fileSize int64, originalName, destDir, subDir string, maxSizeBytes int64) (string, error) {
	if !allowedSubDirs[subDir] {
		return "", fmt.Errorf("非法的子目录: %s", subDir)
	}

	limitedReader := io.LimitReader(fileHeader, maxSizeBytes+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	if int64(len(data)) > maxSizeBytes {
		return "", fmt.Errorf("文件大小超过限制（最大 %d MB）", maxSizeBytes/1024/1024)
	}

	detected := http.DetectContentType(data)
	if !allowedMIMETypes[detected] {
		return "", fmt.Errorf("不支持的文件类型: %s，仅支持 JPEG/PNG/GIF/WebP", detected)
	}

	ext, ok := mimeToExt[detected]
	if !ok {
		ext = ".jpg"
	}

	name, err := generateFilename()
	if err != nil {
		return "", err
	}
	filename := name + ext
	dir := filepath.Join(destDir, subDir)

	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("获取目录路径失败: %w", err)
	}
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("获取目标路径失败: %w", err)
	}
	if !strings.HasPrefix(absDir, absDest) {
		return "", fmt.Errorf("非法的目录路径")
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("保存文件失败: %w", err)
	}

	return filepath.Join(subDir, filename), nil
}

func generateFilename() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("生成文件名失败: %w", err)
	}
	return hex.EncodeToString(b), nil
}
