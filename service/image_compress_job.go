package service

import (
	"api/dao"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"strings"
	"sync"
	"time"
)

// CompressJobStatus 表示压缩任务状态。
type CompressJobStatus string

const (
	CompressJobPending CompressJobStatus = "pending"
	CompressJobRunning CompressJobStatus = "running"
	CompressJobSuccess CompressJobStatus = "success"
	CompressJobFailed  CompressJobStatus = "failed"
)

// CompressProgress 表示通过 SSE 推送给前端的进度事件。
type CompressProgress struct {
	JobID             string            `json:"job_id"`
	Status            CompressJobStatus `json:"status"`
	Current           int               `json:"current"`            // 当前已处理图片数
	Total             int               `json:"total"`              // 总图片数
	OriginalTotal     int64             `json:"original_total"`     // 当前累计原始总大小
	CompressedTotal   int64             `json:"compressed_total"`   // 当前累计压缩后总大小
	CompressedPercent float64           `json:"compressed_percent"` // 当前压缩后 / 原始 * 100
	ReductionPercent  float64           `json:"reduction_percent"`  // 当前缩小百分比 = 100 - CompressedPercent
	TarURL            string            `json:"tar_url,omitempty"`  // 完成后返回的压缩包下载地址（历史字段名，实际为 zip）
	TarPath           string            `json:"tar_path,omitempty"` // 完成后返回的相对路径（zip）
	Error             string            `json:"error,omitempty"`    // 失败时的错误信息
	Done              bool              `json:"done"`               // 是否任务结束（成功或失败）
}

// CompressJob 表示一次压缩任务。
type CompressJob struct {
	ID        string
	Files     []*multipart.FileHeader
	Quality   int
	CreatedAt time.Time

	mu              sync.Mutex
	Status          CompressJobStatus
	TarPath         string
	TarURL          string
	Results         []ImageCompressResult
	Err             error
	progressCh      chan CompressProgress
	originalTotal   int64
	compressedTotal int64
}

var (
	compressJobs   = make(map[string]*CompressJob)
	compressJobsMu sync.Mutex
)

// StartCompressJob 创建并启动一个异步压缩任务。
func StartCompressJob(files []*multipart.FileHeader, quality int, baseURL string) (*CompressJob, error) {
	if len(files) == 0 {
		return nil, fmt.Errorf("没有需要压缩的图片")
	}
	if len(files) > maxImagesPerRequest {
		return nil, fmt.Errorf("单次最多压缩 %d 张图片", maxImagesPerRequest)
	}

	// 质量边界与同步接口相同
	if quality <= 0 {
		quality = defaultJPEGQuality
	}
	if quality < minJPEGQuality {
		quality = minJPEGQuality
	}
	if quality > maxJPEGQuality {
		quality = maxJPEGQuality
	}

	id := GenerateFileName("compress-job") // 复用文件名生成逻辑，保证全局唯一
	job := &CompressJob{
		ID:         id,
		Files:      files,
		Quality:    quality,
		CreatedAt:  time.Now(),
		Status:     CompressJobPending,
		progressCh: make(chan CompressProgress, 16),
	}

	compressJobsMu.Lock()
	compressJobs[id] = job
	compressJobsMu.Unlock()

	// 记录任务创建信息到数据库（总图片数和初始状态）
	_ = dao.CreateImageCompressJob(id, len(files), string(CompressJobPending))

	// 启动后台协程执行压缩
	go job.run(baseURL)

	return job, nil
}

// GetCompressJob 根据 ID 获取任务。
func GetCompressJob(id string) (*CompressJob, bool) {
	compressJobsMu.Lock()
	defer compressJobsMu.Unlock()
	job, ok := compressJobs[id]
	return job, ok
}

// ProgressChan 返回只读进度通道，用于 SSE 推送。
func (j *CompressJob) ProgressChan() <-chan CompressProgress {
	return j.progressCh
}

// run 按顺序压缩每一张图片，并在每张完成后推送一次进度事件。
func (j *CompressJob) run(baseURL string) {
	defer close(j.progressCh)

	j.mu.Lock()
	j.Status = CompressJobRunning
	j.mu.Unlock()

	// 更新数据库中的任务状态为 running
	_ = dao.FinishImageCompressJob(j.ID, string(CompressJobRunning), 0, 0, "", "")

	total := len(j.Files)
	dataBufs := make([][]byte, total)
	results := make([]ImageCompressResult, 0, total)

	for idx, fh := range j.Files {
		res, data, err := compressSingleImage(fh, j.Quality)
		if err != nil {
			j.mu.Lock()
			j.Status = CompressJobFailed
			j.Err = err
			j.mu.Unlock()

			// 持久化失败状态
			_ = dao.FinishImageCompressJob(j.ID, string(CompressJobFailed), j.originalTotal, j.compressedTotal, "", err.Error())

			j.progressCh <- CompressProgress{
				JobID:   j.ID,
				Status:  CompressJobFailed,
				Error:   err.Error(),
				Done:    true,
				Total:   total,
				Current: idx,
			}
			return
		}
		dataBufs[idx] = data
		results = append(results, res)

		j.originalTotal += res.OriginalSize
		j.compressedTotal += res.CompressedSize

		var compressedPercent, reductionPercent float64
		if j.originalTotal > 0 && j.compressedTotal >= 0 {
			compressedPercent = float64(j.compressedTotal) * 100.0 / float64(j.originalTotal)
			reductionPercent = 100.0 - compressedPercent
		}

		j.progressCh <- CompressProgress{
			JobID:             j.ID,
			Status:            CompressJobRunning,
			Current:           idx + 1,
			Total:             total,
			OriginalTotal:     j.originalTotal,
			CompressedTotal:   j.compressedTotal,
			CompressedPercent: compressedPercent,
			ReductionPercent:  reductionPercent,
			Done:              false,
		}
	}

	// 全部单张压缩完成后，写入 zip 包
	tarPath, err := writeImagesToZip(j.Files, dataBufs)
	if err != nil {
		j.mu.Lock()
		j.Status = CompressJobFailed
		j.Err = err
		j.mu.Unlock()

		_ = dao.FinishImageCompressJob(j.ID, string(CompressJobFailed), j.originalTotal, j.compressedTotal, "", err.Error())

		j.progressCh <- CompressProgress{
			JobID:   j.ID,
			Status:  CompressJobFailed,
			Error:   err.Error(),
			Done:    true,
			Total:   total,
			Current: total,
		}
		return
	}

	relativePath := strings.TrimPrefix(tarPath, "./")
	downloadURL := fmt.Sprintf("/api/tools/image-compress/download?job_id=%s", j.ID)

	j.mu.Lock()
	j.Status = CompressJobSuccess
	j.TarPath = relativePath
	j.TarURL = downloadURL
	j.Results = results
	j.mu.Unlock()

	// 成功时，更新任务记录并累加到全局统计
	_ = dao.FinishImageCompressJob(j.ID, string(CompressJobSuccess), j.originalTotal, j.compressedTotal, relativePath, "")
	_ = dao.AddImageCompressStats(int64(total), j.originalTotal, j.compressedTotal)

	var compressedPercent, reductionPercent float64
	if j.originalTotal > 0 && j.compressedTotal >= 0 {
		compressedPercent = float64(j.compressedTotal) * 100.0 / float64(j.originalTotal)
		reductionPercent = 100.0 - compressedPercent
	}

	j.progressCh <- CompressProgress{
		JobID:             j.ID,
		Status:            CompressJobSuccess,
		Current:           total,
		Total:             total,
		OriginalTotal:     j.originalTotal,
		CompressedTotal:   j.compressedTotal,
		CompressedPercent: compressedPercent,
		ReductionPercent:  reductionPercent,
		TarURL:            downloadURL,
		TarPath:           relativePath,
		Done:              true,
	}
}

// EncodeProgressEvent 将进度结构编码为 JSON 字节，供 SSE 使用。
func EncodeProgressEvent(p CompressProgress) []byte {
	b, _ := json.Marshal(p)
	return b
}
