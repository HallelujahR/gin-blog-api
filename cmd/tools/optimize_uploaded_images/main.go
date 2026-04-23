package main

import (
	"api/service"
	"flag"
	"fmt"
	"log"
)

func main() {
	var (
		dir             = flag.String("dir", service.ImageUploadDir, "需要扫描的图片目录")
		backupDir       = flag.String("backup-dir", "./uploads/image_backups", "原图备份目录")
		quality         = flag.Int("quality", 80, "JPEG 压缩质量，范围 10-95")
		minSavingsBytes = flag.Int64("min-savings-bytes", 4*1024, "最少节省多少字节才覆盖原图")
		minSavingsRatio = flag.Float64("min-savings-ratio", 0.08, "最少节省比例才覆盖原图")
		noBackup        = flag.Bool("no-backup", false, "是否禁用原图备份")
	)
	flag.Parse()

	report, err := service.OptimizeLocalImages(service.LocalImageOptimizeOptions{
		RootDir:         *dir,
		BackupDir:       *backupDir,
		Quality:         *quality,
		MinSavingsBytes: *minSavingsBytes,
		MinSavingsRatio: *minSavingsRatio,
		KeepBackup:      !*noBackup,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("扫描文件: %d\n", report.ScannedFiles)
	fmt.Printf("可处理图片: %d\n", report.SupportedFiles)
	fmt.Printf("成功压缩: %d\n", report.OptimizedFiles)
	fmt.Printf("跳过文件: %d\n", report.SkippedFiles)
	fmt.Printf("失败文件: %d\n", report.FailedFiles)
	fmt.Printf("原始体积: %.2f MB\n", float64(report.OriginalBytes)/(1<<20))
	fmt.Printf("压缩后体积: %.2f MB\n", float64(report.OptimizedBytes)/(1<<20))
	fmt.Printf("节省体积: %.2f MB\n", float64(report.SavedBytes)/(1<<20))
}
