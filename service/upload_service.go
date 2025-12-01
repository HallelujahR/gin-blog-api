package service

import (
	"api/configs"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// 上传文件存储目录
const (
	UploadDir          = "./uploads"
	ImageUploadDir     = "./uploads/images"
	FileUploadDir      = "./uploads/files"
	CompressedTempDir  = "./uploads/compressed_tmp" // 临时压缩结果目录，仅保留短时间用于下载
	PublicURL          = "/uploads"                 // 公开访问URL前缀
	compressKeepPeriod = time.Hour                  // 压缩结果保留时长
)

// 初始化上传目录
func InitUploadDirs() error {
	dirs := []string{UploadDir, ImageUploadDir, FileUploadDir, CompressedTempDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败 %s: %v", dir, err)
		}
	}
	// 启动后台清理协程，定期清理过期的临时压缩包，避免磁盘被占满。
	go cleanupCompressedTempFiles()
	return nil
}

// cleanupCompressedTempFiles 周期性扫描临时压缩目录，将超过保留时间的 tar 包删除。
func cleanupCompressedTempFiles() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		_ = filepath.Walk(CompressedTempDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if info.IsDir() {
				return nil
			}
			// 仅处理 .tar 文件
			if strings.ToLower(filepath.Ext(info.Name())) != ".tar" {
				return nil
			}
			if time.Since(info.ModTime()) > compressKeepPeriod {
				_ = os.Remove(path)
			}
			return nil
		})
	}
}

// CompressKeepPeriod 暴露压缩结果的保留时间，用于 API 返回给前端。
func CompressKeepPeriod() time.Duration {
	return compressKeepPeriod
}

// 生成唯一文件名
func GenerateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	name := strings.TrimSuffix(originalName, ext)
	// 清理文件名，只保留字母数字和短横线
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, " ", "-")

	// 生成唯一ID
	uniqueID := uuid.New().String()[:8]
	timestamp := time.Now().Format("20060102")

	return fmt.Sprintf("%s-%s-%s%s", timestamp, uniqueID, name, ext)
}

// 保存base64编码的图片数据
func SaveBase64Image(base64Data string, filename string) (string, error) {
	// 检测并移除data URL前缀（如：data:image/jpeg;base64,）
	var imageData []byte
	var err error

	if strings.HasPrefix(base64Data, "data:image/") {
		// 包含data URL前缀
		parts := strings.Split(base64Data, ",")
		if len(parts) != 2 {
			return "", fmt.Errorf("无效的base64数据格式")
		}
		imageData, err = base64.StdEncoding.DecodeString(parts[1])
	} else {
		// 纯base64数据
		imageData, err = base64.StdEncoding.DecodeString(base64Data)
	}

	if err != nil {
		return "", fmt.Errorf("base64解码失败: %v", err)
	}

	// 如果没有提供文件名，根据数据生成
	if filename == "" {
		// 检测图片格式
		ext := ".jpg" // 默认
		if len(imageData) > 4 {
			// 检测PNG: 89 50 4E 47
			if imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47 {
				ext = ".png"
			} else if len(imageData) > 2 {
				// 检测JPEG: FF D8
				if imageData[0] == 0xFF && imageData[1] == 0xD8 {
					ext = ".jpg"
				} else if imageData[0] == 0x47 && imageData[1] == 0x49 && len(imageData) > 3 && imageData[2] == 0x46 {
					// GIF: 47 49 46
					ext = ".gif"
				} else if len(imageData) > 12 && string(imageData[0:4]) == "RIFF" && string(imageData[8:12]) == "WEBP" {
					// WEBP
					ext = ".webp"
				}
			}
		}
		filename = GenerateFileName("image" + ext)
	} else {
		// 确保文件名有扩展名
		if filepath.Ext(filename) == "" {
			// 根据数据检测扩展名
			ext := ".jpg"
			if len(imageData) > 4 {
				if imageData[0] == 0x89 && imageData[1] == 0x50 && imageData[2] == 0x4E && imageData[3] == 0x47 {
					ext = ".png"
				} else if len(imageData) > 2 && imageData[0] == 0xFF && imageData[1] == 0xD8 {
					ext = ".jpg"
				} else if len(imageData) > 2 && imageData[0] == 0x47 && imageData[1] == 0x49 {
					ext = ".gif"
				}
			}
			filename = filename + ext
		}
		filename = GenerateFileName(filename)
	}

	// 确定保存目录
	saveDir := ImageUploadDir

	// 创建保存路径
	savePath := filepath.Join(saveDir, filename)

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 写入文件内容
	if _, err = dst.Write(imageData); err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	// 返回相对路径（用于URL）
	relativePath := strings.TrimPrefix(savePath, "./")
	return relativePath, nil
}

// 保存上传的文件
func SaveUploadedFile(file *multipart.FileHeader, filename string) (string, error) {
	// 打开上传的文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %v", err)
	}
	defer src.Close()

	// 确定保存目录
	ext := strings.ToLower(filepath.Ext(filename))
	var saveDir string
	if isImage(ext) {
		saveDir = ImageUploadDir
	} else {
		saveDir = FileUploadDir
	}

	// 创建保存路径
	savePath := filepath.Join(saveDir, filename)

	// 创建目标文件
	dst, err := os.Create(savePath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("保存文件失败: %v", err)
	}

	// 返回相对路径（用于URL）
	relativePath := strings.TrimPrefix(savePath, "./")
	return relativePath, nil
}

// 获取文件访问URL（相对路径）
func GetFileURL(filePath string) string {
	// 如果已经是完整URL，直接返回
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return filePath
	}

	// 确保路径以/开头
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	// 返回相对路径
	return filePath
}

// 获取完整的文件访问URL（从配置文件读取BaseURL）
func GetFullFileURL(filePath string) string {
	// 如果已经是完整URL，直接返回
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		return filePath
	}

	// 确保路径以/开头
	if !strings.HasPrefix(filePath, "/") {
		filePath = "/" + filePath
	}

	// 从配置获取基础URL
	baseURL := strings.TrimSuffix(configs.GetBaseURL(), "/")
	return baseURL + filePath
}

// 删除文件
func DeleteFile(filePath string) error {
	// 如果传入的是URL，提取路径
	if strings.HasPrefix(filePath, "http://") || strings.HasPrefix(filePath, "https://") {
		// 从URL中提取路径部分
		parts := strings.Split(filePath, "/uploads/")
		if len(parts) > 1 {
			filePath = "./uploads/" + parts[1]
		}
	} else if !strings.HasPrefix(filePath, "./") {
		// 如果是相对路径，添加前缀
		if !strings.HasPrefix(filePath, "uploads/") {
			filePath = "./uploads/" + filePath
		} else {
			filePath = "./" + filePath
		}
	}

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}

	// 删除文件
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// 列出已上传的文件
func ListUploadedFiles(fileType string, page, pageSize int) ([]map[string]interface{}, int, error) {
	var searchDirs []string

	switch fileType {
	case "image":
		searchDirs = []string{ImageUploadDir}
	case "file":
		searchDirs = []string{FileUploadDir}
	default:
		searchDirs = []string{ImageUploadDir, FileUploadDir}
	}

	var allFiles []map[string]interface{}

	for _, dir := range searchDirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 跳过错误
			}
			if !info.IsDir() {
				relativePath := strings.TrimPrefix(path, "./")
				fileURL := GetFileURL(relativePath)

				allFiles = append(allFiles, map[string]interface{}{
					"name":     info.Name(),
					"path":     relativePath,
					"url":      fileURL,
					"size":     info.Size(),
					"modified": info.ModTime(),
				})
			}
			return nil
		})
		if err != nil {
			return nil, 0, err
		}
	}

	total := len(allFiles)

	// 分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}

	if start < end {
		return allFiles[start:end], total, nil
	}
	return []map[string]interface{}{}, total, nil
}

// 判断是否为图片
func isImage(ext string) bool {
	imageExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
		".svg":  true,
	}
	return imageExts[ext]
}
