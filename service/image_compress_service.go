package service

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	_ "image/gif"  // 注册 GIF 解码器
	_ "image/jpeg" // 注册 JPEG 解码器
	_ "image/png"  // 注册 PNG 解码器
)

// ImageCompressResult 描述单张图片的压缩结果与“进度百分比”（压缩率）。
// Progress 表示压缩后的大小占原始大小的百分比，数值越小压缩效果越好。
type ImageCompressResult struct {
	Name           string  `json:"name"`
	OriginalSize   int64   `json:"original_size"`
	CompressedSize int64   `json:"compressed_size"`
	Progress       float64 `json:"progress"` // 压缩后大小 / 原始大小 * 100
}

const (
	maxImagesPerRequest   = 10       // 单次最多处理 10 张图片
	maxSingleImageSize    = 15 << 20 // 单张图片最大 15MB（不低于 5MB 的要求）
	defaultJPEGQuality    = 80       // 默认压缩质量
	minJPEGQuality        = 10       // 允许的最小压缩质量
	maxJPEGQuality        = 95       // 允许的最大压缩质量
	maxConcurrentCompress = 4        // 单次请求内并发压缩协程数
)

// CompressImages 并发压缩多张图片，返回 zip 包路径和每张图片的压缩结果。
// 该函数不持久化原始文件，只在内存中处理图片数据。
func CompressImages(files []*multipart.FileHeader, quality int) (string, []ImageCompressResult, error) {
	if len(files) == 0 {
		return "", nil, fmt.Errorf("没有需要压缩的图片")
	}
	if len(files) > maxImagesPerRequest {
		return "", nil, fmt.Errorf("单次最多压缩 %d 张图片", maxImagesPerRequest)
	}

	if quality <= 0 {
		quality = defaultJPEGQuality
	}
	if quality < minJPEGQuality {
		quality = minJPEGQuality
	}
	if quality > maxJPEGQuality {
		quality = maxJPEGQuality
	}

	results := make([]ImageCompressResult, len(files))
	dataBufs := make([][]byte, len(files))

	var wg sync.WaitGroup
	errCh := make(chan error, len(files))
	sem := make(chan struct{}, maxConcurrentCompress)

	for i, fh := range files {
		i, fh := i, fh
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			res, data, err := compressSingleImage(fh, quality)
			if err != nil {
				errCh <- fmt.Errorf("%s: %w", fh.Filename, err)
				return
			}
			results[i] = res
			dataBufs[i] = data
		}()
	}

	wg.Wait()
	close(errCh)
	if err, ok := <-errCh; ok {
		// 只返回第一条错误信息，前端可提示重试或减少图片数量
		return "", nil, err
	}

	// 将所有压缩后的图片打包为 zip 文件，放置在临时目录，供 1 小时内下载。
	tarPath, err := writeImagesToZip(files, dataBufs)
	if err != nil {
		return "", nil, err
	}
	return tarPath, results, nil
}

// compressSingleImage 对单张图片进行安全校验并压缩，返回压缩结果和压缩后字节数据。
func compressSingleImage(fh *multipart.FileHeader, quality int) (ImageCompressResult, []byte, error) {
	if fh.Size > maxSingleImageSize {
		return ImageCompressResult{}, nil, fmt.Errorf("图片过大，单张最大允许 %.1fMB", float64(maxSingleImageSize)/(1<<20))
	}

	// 打开上传文件流
	file, err := fh.Open()
	if err != nil {
		return ImageCompressResult{}, nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	// 读取前 512 字节进行 MIME 类型探测
	headerBuf := make([]byte, 512)
	n, _ := io.ReadFull(file, headerBuf)
	contentType := http.DetectContentType(headerBuf[:n])
	if !strings.HasPrefix(contentType, "image/") {
		return ImageCompressResult{}, nil, fmt.Errorf("非法文件类型: %s", contentType)
	}

	// 重新打开文件用于解码（上面的读取已经移动了指针）
	_ = file.Close()
	file, err = fh.Open()
	if err != nil {
		return ImageCompressResult{}, nil, fmt.Errorf("重新打开文件失败: %w", err)
	}
	defer file.Close()

	// 限制最大读取大小，防止畸形文件占用过多内存
	limited := io.LimitReader(file, maxSingleImageSize+1)

	// 尝试解码图片，若解码失败视为“包含恶意代码或非法图片”
	img, format, err := image.Decode(limited)
	if err != nil {
		return ImageCompressResult{}, nil, fmt.Errorf("图片解码失败，可能为损坏或恶意文件: %w", err)
	}

	originalSize := fh.Size

	// 根据图片格式进行压缩编码
	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return ImageCompressResult{}, nil, fmt.Errorf("JPEG 压缩失败: %w", err)
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, img); err != nil {
			return ImageCompressResult{}, nil, fmt.Errorf("PNG 压缩失败: %w", err)
		}
	case "gif":
		if err := gif.Encode(&buf, img, nil); err != nil {
			return ImageCompressResult{}, nil, fmt.Errorf("GIF 压缩失败: %w", err)
		}
	default:
		return ImageCompressResult{}, nil, fmt.Errorf("暂不支持的图片格式: %s", format)
	}

	compressedSize := int64(buf.Len())
	progress := 100.0
	if originalSize > 0 {
		progress = float64(compressedSize) * 100.0 / float64(originalSize)
	}

	res := ImageCompressResult{
		Name:           fh.Filename,
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
		Progress:       round2(progress),
	}
	return res, buf.Bytes(), nil
}

// writeImagesToZip 将多个压缩后的图片字节写入一个 zip 包文件。
func writeImagesToZip(files []*multipart.FileHeader, dataBufs [][]byte) (string, error) {
	if len(files) != len(dataBufs) {
		return "", fmt.Errorf("内部错误: 文件与数据长度不一致")
	}

	if err := os.MkdirAll(CompressedTempDir, 0o755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	timestamp := time.Now().Format("20060102-150405")
	tarName := fmt.Sprintf("compressed-%s.zip", timestamp)
	tarPath := filepath.Join(CompressedTempDir, tarName)

	f, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("创建 zip 文件失败: %w", err)
	}
	defer f.Close()

	tw := zip.NewWriter(f)
	defer tw.Close()

	for i, fh := range files {
		data := dataBufs[i]
		if data == nil {
			continue
		}
		hdr := &zip.FileHeader{
			Name:     safeArchiveName(fh.Filename),
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		writer, err := tw.CreateHeader(hdr)
		if err != nil {
			return "", fmt.Errorf("写入 zip 头失败: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			return "", fmt.Errorf("写入 zip 内容失败: %w", err)
		}
	}

	if err := tw.Close(); err != nil {
		return "", fmt.Errorf("关闭 zip 写入器失败: %w", err)
	}
	if err := f.Close(); err != nil {
		return "", fmt.Errorf("关闭 zip 文件失败: %w", err)
	}
	return tarPath, nil
}

// safeArchiveName 清理文件名，避免目录穿越等问题。
func safeArchiveName(name string) string {
	name = filepath.Base(name)
	name = strings.TrimSpace(name)
	if name == "" {
		name = "image"
	}
	return name
}

// round2 保留两位小数。
func round2(v float64) float64 {
	return math.Round(v*100) / 100
}
