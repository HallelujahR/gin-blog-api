package configs

import "os"

var (
	// BaseURL API服务的基础URL，用于生成完整的图片访问地址
	// 默认值：http://localhost:8080（开发环境）
	// 生产环境可通过环境变量API_BASE_URL设置
	BaseURL = "http://localhost:8080"
)

func init() {
	// 从环境变量读取BaseURL（如果设置了）
	if apiBaseURL := os.Getenv("API_BASE_URL"); apiBaseURL != "" {
		BaseURL = apiBaseURL
	}
}

// GetBaseURL 获取基础URL
func GetBaseURL() string {
	return BaseURL
}

// SetBaseURL 设置基础URL（例如：http://localhost:8080 或 https://api.example.com）
func SetBaseURL(url string) {
	if url != "" {
		BaseURL = url
	}
}
