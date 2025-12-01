package admin

import (
	"api/configs"
	"api/dao"
	"api/service"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

// （同步压缩接口已移除，仅保留异步 + SSE 通信）

// StartCompressJob 启动异步压缩任务，返回 job_id，前端可通过 SSE 订阅进度。
// 请求方式与 CompressImages 类似：multipart/form-data，字段 images / file，quality。
func StartCompressJob(c *gin.Context) {
	const maxTotalSize = 100 << 20
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxTotalSize)

	qualityStr := c.DefaultPostForm("quality", "80")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil {
		quality = 80
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "表单解析失败: " + err.Error()})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		files = form.File["file"]
	}
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有上传任何图片"})
		return
	}

	baseURL := configs.GetBaseURL()
	job, err := service.StartCompressJob(files, quality, baseURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "创建压缩任务失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"job_id":    job.ID,
		"status":    job.Status,
		"createdAt": job.CreatedAt.Format(time.RFC3339),
	})
}

// StreamCompressProgress 通过 SSE 推送压缩进度。
// GET /api/admin/upload/compress/stream?job_id=xxx
func StreamCompressProgress(c *gin.Context) {
	jobID := c.Query("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 job_id"})
		return
	}

	job, ok := service.GetCompressJob(jobID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "压缩任务不存在"})
		return
	}

	// SSE 头部设置
	w := c.Writer
	header := w.Header()
	header.Set("Content-Type", "text/event-stream")
	header.Set("Cache-Control", "no-cache")
	header.Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "当前服务器不支持 SSE"})
		return
	}

	ctx := c.Request.Context()
	ch := job.ProgressChan()

	// 立即刷新一次，让浏览器建立连接
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			return
		case p, ok := <-ch:
			if !ok {
				return
			}
			data := service.EncodeProgressEvent(p)
			// SSE 格式: data: <json>\n\n
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			if p.Done {
				return
			}
		}
	}
}

// DownloadCompressResult 下载压缩结果文件，强制浏览器以附件形式保存。
// GET /api/tools/image-compress/download?job_id=xxx
func DownloadCompressResult(c *gin.Context) {
	jobID := c.Query("job_id")
	if jobID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 job_id"})
		return
	}

	job, err := dao.GetImageCompressJob(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "压缩任务不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务失败: " + err.Error()})
		}
		return
	}
	if job.TarPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "任务尚未生成压缩包"})
		return
	}

	filePath := job.TarPath
	if strings.HasPrefix(filePath, "/") {
		filePath = "." + filePath
	} else if !strings.HasPrefix(filePath, "./") {
		filePath = "./" + filePath
	}

	if info, statErr := os.Stat(filePath); statErr != nil {
		if os.IsNotExist(statErr) {
			c.JSON(http.StatusNotFound, gin.H{"error": "压缩结果已过期或被删除"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败: " + statErr.Error()})
		}
		return
	} else if info.IsDir() {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "内部错误: 指向目录而非文件"})
		return
	}

	filename := filepath.Base(filePath)
	c.Header("Content-Type", "application/zip")
	c.FileAttachment(filePath, filename)
}

// GetCompressStats 返回累计任务数量、累计成功处理图片数量以及累计节省空间大小。
// GET /api/admin/upload/compress/stats
func GetCompressStats(c *gin.Context) {
	totalJobs, err := dao.CountImageCompressJobs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询任务数量失败: " + err.Error()})
		return
	}

	stats, err := dao.GetImageCompressStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询累计统计失败: " + err.Error()})
		return
	}

	savedBytes := stats.TotalOriginalBytes - stats.TotalCompressedBytes
	if savedBytes < 0 {
		savedBytes = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"total_jobs":             totalJobs,                  // 累计任务数量
		"total_images":           stats.TotalImages,          // 累计成功处理的图片张数
		"total_original_bytes":   stats.TotalOriginalBytes,   // 累计原始大小
		"total_compressed_bytes": stats.TotalCompressedBytes, // 累计压缩后大小
		"saved_bytes":            savedBytes,                 // 累计节省的空间大小
	})
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
