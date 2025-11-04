package admin

import (
	"api/service"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// 上传文件（单文件）
func UploadFile(c *gin.Context) {
	// 限制文件大小：10MB
	const maxSize = 10 << 20 // 10MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败: " + err.Error()})
		return
	}

	// 验证文件类型
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".zip":  true,
		".rar":  true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("不支持的文件类型: %s", ext),
		})
		return
	}

	// 生成唯一文件名
	filename := service.GenerateFileName(file.Filename)

	// 保存文件
	filePath, err := service.SaveUploadedFile(file, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败: " + err.Error()})
		return
	}

	// 生成完整访问URL
	fileURL := service.GetFullFileURL(filePath)

	// 返回文件信息
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"path":     filePath,
		"filename": filename,
		"original": file.Filename,
		"size":     file.Size,
		"type":     file.Header.Get("Content-Type"),
	})
}

// 上传图片（专门用于图片，自动压缩和格式转换）
func UploadImage(c *gin.Context) {
	// 限制文件大小：5MB
	const maxSize = 5 << 20 // 5MB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	// 获取上传的文件
	file, err := c.FormFile("image")
	if err != nil {
		// 尝试其他字段名
		file, err = c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片上传失败: " + err.Error()})
			return
		}
	}

	// 验证是否为图片
	allowedImageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedImageExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("不是有效的图片格式: %s，支持 jpg, png, gif, webp, svg", ext),
		})
		return
	}

	// 生成唯一文件名
	filename := service.GenerateFileName(file.Filename)

	// 保存文件
	filePath, err := service.SaveUploadedFile(file, filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "图片保存失败: " + err.Error()})
		return
	}

	// 生成完整访问URL
	fileURL := service.GetFullFileURL(filePath)

	// 返回图片信息
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"path":     filePath,
		"filename": filename,
		"original": file.Filename,
		"size":     file.Size,
		"type":     file.Header.Get("Content-Type"),
	})
}

// 批量上传文件
func UploadFiles(c *gin.Context) {
	// 限制总大小：50MB
	const maxSize = 50 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "表单解析失败: " + err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有上传文件"})
		return
	}

	var results []gin.H
	var errors []string

	for _, file := range files {
		// 验证文件类型
		allowedExts := map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
			".webp": true, ".svg": true, ".pdf": true,
			".doc": true, ".docx": true, ".zip": true, ".rar": true,
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedExts[ext] {
			errors = append(errors, fmt.Sprintf("%s: 不支持的文件类型", file.Filename))
			continue
		}

		// 生成唯一文件名
		filename := service.GenerateFileName(file.Filename)

		// 保存文件
		filePath, err := service.SaveUploadedFile(file, filename)
		if err != nil {
			errors = append(errors, fmt.Sprintf("%s: %s", file.Filename, err.Error()))
			continue
		}

		// 生成完整访问URL
		fileURL := service.GetFullFileURL(filePath)

		results = append(results, gin.H{
			"url":      fileURL,
			"path":     filePath,
			"filename": filename,
			"original": file.Filename,
			"size":     file.Size,
			"type":     file.Header.Get("Content-Type"),
		})
	}

	response := gin.H{
		"success": len(results),
		"total":   len(files),
		"files":   results,
	}
	if len(errors) > 0 {
		response["errors"] = errors
	}

	c.JSON(http.StatusOK, response)
}

// 删除文件
func DeleteFile(c *gin.Context) {
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少文件路径参数"})
		return
	}

	err := service.DeleteFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// 获取文件列表（可选功能）
func ListFiles(c *gin.Context) {
	fileType := c.Query("type") // image, file, all
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if ps := c.Query("page_size"); ps != "" {
		fmt.Sscanf(ps, "%d", &pageSize)
	}

	files, total, err := service.ListUploadedFiles(fileType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件列表失败: " + err.Error()})
		return
	}

	// 确保返回的文件URL都是完整URL
	for i := range files {
		if url, ok := files[i]["url"].(string); ok && url != "" {
			files[i]["url"] = service.GetFullFileURL(url)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"files":     files,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}
