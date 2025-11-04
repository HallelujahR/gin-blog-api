# 图片URL路径简化说明

## 简化内容

已简化图片完整路径获取逻辑，直接从配置文件读取域名/地址，不再需要从请求中获取。

## 主要改动

### 1. 配置模块 (`configs/config.go`)

**之前**：需要手动调用 `SetBaseURL()` 设置  
**现在**：默认值 `http://localhost:8080`，可直接使用

```go
var (
    // BaseURL API服务的基础URL，默认值：http://localhost:8080（开发环境）
    BaseURL = "http://localhost:8080"
)

// SetBaseURL 设置基础URL（生产环境可调用此方法修改）
func SetBaseURL(url string) {
    if url != "" {
        BaseURL = url
    }
}
```

### 2. 服务层 (`service/upload_service.go`)

**之前**：需要传入 `gin.Context`，从请求中获取 scheme 和 host  
**现在**：直接传入文件路径，从配置读取 BaseURL

```go
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
```

### 3. 控制器调用简化

**之前**：
```go
getFullURL := service.GetFullFileURL(c)
post.CoverImage = getFullURL(post.CoverImage)
```

**现在**：
```go
post.CoverImage = service.GetFullFileURL(post.CoverImage)
```

## 使用方式

### 开发环境（默认）

无需任何配置，默认使用 `http://localhost:8080`

```go
// 直接使用，自动使用默认值
url := service.GetFullFileURL("/uploads/images/xxx.jpg")
// 返回: http://localhost:8080/uploads/images/xxx.jpg
```

### 生产环境

在应用启动时设置 BaseURL：

```go
import "api/configs"

func main() {
    // 设置生产环境的基础URL
    configs.SetBaseURL("https://api.example.com")
    
    // ... 其他初始化代码
    initializer.Run()
}
```

或使用环境变量：

```go
func init() {
    if baseURL := os.Getenv("API_BASE_URL"); baseURL != "" {
        configs.SetBaseURL(baseURL)
    }
}
```

## 优势

1. ✅ **简化调用**：不需要传入 context，直接传入文件路径
2. ✅ **统一配置**：所有图片URL都从同一个配置读取
3. ✅ **易于维护**：修改域名只需改配置，无需改代码
4. ✅ **性能更好**：不需要每次请求都解析请求头
5. ✅ **代码更清晰**：移除了复杂的请求解析逻辑

## 修复范围

### 后台管理接口
- ✅ 文章创建/更新/详情/列表
- ✅ 文件上传/列表

### 前端用户接口
- ✅ 文章详情
- ✅ 文章列表

## 示例

### 创建文章（带图片）

**请求**：
```bash
POST /api/admin/posts
Content-Type: multipart/form-data

image: [文件]
title: 测试文章
```

**响应**：
```json
{
  "post": {
    "id": 1,
    "title": "测试文章",
    "cover_image": "http://localhost:8080/uploads/images/20251103-xxx.jpg"
  }
}
```

### 获取文章列表

**请求**：
```bash
GET /api/posts?page=1&size=10
```

**响应**：
```json
{
  "posts": [
    {
      "id": 1,
      "title": "文章标题",
      "cover_image": "http://localhost:8080/uploads/images/xxx.jpg"
    }
  ]
}
```

## 配置说明

### 开发环境

默认值：`http://localhost:8080`，无需配置

### 生产环境

**方式1：代码配置**
```go
configs.SetBaseURL("https://api.example.com")
```

**方式2：环境变量**
```bash
export API_BASE_URL=https://api.example.com
```

然后在代码中：
```go
if baseURL := os.Getenv("API_BASE_URL"); baseURL != "" {
    configs.SetBaseURL(baseURL)
}
```

## 总结

✅ **简化完成**：代码更简洁，调用更方便  
✅ **统一配置**：所有图片URL从配置读取  
✅ **易于维护**：修改域名只需改配置  
✅ **向后兼容**：如果已经是完整URL，不会重复转换  

现在图片路径获取逻辑已大大简化！🎉

