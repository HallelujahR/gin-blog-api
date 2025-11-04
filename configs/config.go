package configs

var (
	// BaseURL API服务的基础URL，用于生成完整的图片访问地址
	// 默认值：http://localhost:8080（开发环境）
	BaseURL = "http://localhost:8080"
)

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
