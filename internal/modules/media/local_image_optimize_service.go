package media

import (
	"bytes"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type LocalImageOptimizeOptions struct {
	RootDir         string
	BackupDir       string
	Quality         int
	MinSavingsBytes int64
	MinSavingsRatio float64
	KeepBackup      bool
}

type LocalImageOptimizeReport struct {
	ScannedFiles   int
	SupportedFiles int
	OptimizedFiles int
	SkippedFiles   int
	FailedFiles    int
	OriginalBytes  int64
	OptimizedBytes int64
	SavedBytes     int64
}

func OptimizeLocalImages(opts LocalImageOptimizeOptions) (LocalImageOptimizeReport, error) {
	report := LocalImageOptimizeReport{}

	rootDir := strings.TrimSpace(opts.RootDir)
	if rootDir == "" {
		rootDir = ImageUploadDir
	}
	quality := opts.Quality
	if quality <= 0 {
		quality = defaultJPEGQuality
	}
	if quality < minJPEGQuality {
		quality = minJPEGQuality
	}
	if quality > maxJPEGQuality {
		quality = maxJPEGQuality
	}
	if opts.MinSavingsBytes <= 0 {
		opts.MinSavingsBytes = 4 * 1024
	}
	if opts.MinSavingsRatio <= 0 {
		opts.MinSavingsRatio = 0.08
	}

	info, err := os.Stat(rootDir)
	if err != nil {
		return report, fmt.Errorf("读取图片目录失败: %w", err)
	}
	if !info.IsDir() {
		return report, fmt.Errorf("图片目录不是有效目录: %s", rootDir)
	}

	if opts.KeepBackup {
		backupDir := strings.TrimSpace(opts.BackupDir)
		if backupDir == "" {
			backupDir = filepath.Join(UploadDir, "image_backups")
		}
		opts.BackupDir = backupDir
		if err := os.MkdirAll(backupDir, 0o755); err != nil {
			return report, fmt.Errorf("创建备份目录失败: %w", err)
		}
	}

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			report.FailedFiles++
			return nil
		}
		if info.IsDir() {
			return nil
		}

		report.ScannedFiles++
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if !isOptimizableImageExt(ext) {
			report.SkippedFiles++
			return nil
		}

		report.SupportedFiles++
		report.OriginalBytes += info.Size()

		optimized, err := optimizeLocalImageFile(path, rootDir, opts, quality)
		if err != nil {
			report.FailedFiles++
			return nil
		}
		if optimized.saved {
			report.OptimizedFiles++
			report.OptimizedBytes += optimized.afterSize
			report.SavedBytes += optimized.savedBytes
		} else {
			report.SkippedFiles++
			report.OptimizedBytes += info.Size()
		}
		return nil
	})
	if err != nil {
		return report, fmt.Errorf("扫描图片目录失败: %w", err)
	}

	return report, nil
}

type optimizeFileResult struct {
	saved      bool
	afterSize  int64
	savedBytes int64
}

func optimizeLocalImageFile(path, rootDir string, opts LocalImageOptimizeOptions, quality int) (optimizeFileResult, error) {
	originalData, err := os.ReadFile(path)
	if err != nil {
		return optimizeFileResult{}, fmt.Errorf("读取文件失败: %w", err)
	}
	originalSize := int64(len(originalData))
	if originalSize == 0 {
		return optimizeFileResult{saved: false, afterSize: 0}, nil
	}

	img, format, err := image.Decode(bytes.NewReader(originalData))
	if err != nil {
		return optimizeFileResult{}, fmt.Errorf("解码图片失败: %w", err)
	}

	var buf bytes.Buffer
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality}); err != nil {
			return optimizeFileResult{}, fmt.Errorf("JPEG 重编码失败: %w", err)
		}
	case "png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(&buf, img); err != nil {
			return optimizeFileResult{}, fmt.Errorf("PNG 重编码失败: %w", err)
		}
	case "gif":
		if err := gif.Encode(&buf, img, nil); err != nil {
			return optimizeFileResult{}, fmt.Errorf("GIF 重编码失败: %w", err)
		}
	default:
		return optimizeFileResult{saved: false, afterSize: originalSize}, nil
	}

	newData := buf.Bytes()
	newSize := int64(len(newData))
	savedBytes := originalSize - newSize
	savedRatio := float64(savedBytes) / float64(originalSize)
	if savedBytes < opts.MinSavingsBytes || savedRatio < opts.MinSavingsRatio {
		return optimizeFileResult{saved: false, afterSize: originalSize}, nil
	}

	if opts.KeepBackup {
		relPath, err := filepath.Rel(rootDir, path)
		if err != nil {
			return optimizeFileResult{}, fmt.Errorf("计算备份相对路径失败: %w", err)
		}
		backupPath := filepath.Join(opts.BackupDir, relPath)
		if err := os.MkdirAll(filepath.Dir(backupPath), 0o755); err != nil {
			return optimizeFileResult{}, fmt.Errorf("创建备份子目录失败: %w", err)
		}
		if err := copyFile(path, backupPath); err != nil {
			return optimizeFileResult{}, fmt.Errorf("备份原图失败: %w", err)
		}
	}

	if err := os.WriteFile(path, newData, 0o644); err != nil {
		return optimizeFileResult{}, fmt.Errorf("写回压缩图片失败: %w", err)
	}

	return optimizeFileResult{
		saved:      true,
		afterSize:  newSize,
		savedBytes: savedBytes,
	}, nil
}

func isOptimizableImageExt(ext string) bool {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg", ".png", ".gif":
		return true
	default:
		return false
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Sync()
}
