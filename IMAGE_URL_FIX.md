# 图片URL完整路径修复说明

## 修复内容

已完善所有接口返回的图片路径，确保返回的是完整的可访问URL（包含协议、域名和端口）。

## 修复范围

### 1. 配置模块 (`configs/config.go`)
- 添加了 `BaseURL` 配置项
- 支持手动设置基础URL（如：`http://localhost:8080`）
- 如果未配置，则从请求中自动获取

### 2. 服务层 (`service/upload_service.go`)
- 新增 `GetFullFileURL(c)` 函数，返回一个函数用于生成完整URL
- 自动从请求中获取 `scheme`（http/https）和 `host`
- 支持 `X-Forwarded-Proto` 头（用于代理/负载均衡器场景）
- 如果无法从请求获取，则使用配置的 `BaseURL`
- 如果都没有，则返回相对路径

### 3. 后台管理接口

#### 文章管理 (`controllers/admin/post_controller.go`)
- ✅ **创建文章**：返回的 `cover_image` 是完整URL
- ✅ **更新文章**：返回的 `cover_image` 是完整URL
- ✅ **获取文章详情**：返回的 `cover_image` 是完整URL
- ✅ **文章列表**：列表中所有文章的 `cover_image` 都是完整URL

#### 文件上传 (`controllers/admin/upload_controller.go`)
- ✅ **上传文件**：返回的 `url` 是完整URL
- ✅ **上传图片**：返回的 `url` 是完整URL
- ✅ **批量上传**：返回的所有 `url` 都是完整URL
- ✅ **文件列表**：返回的所有文件 `url` 都是完整URL

### 4. 前端用户接口 (`controllers/post_controller.go`)
- ✅ **获取文章详情**：返回的 `cover_image` 是完整URL
- ✅ **文章列表**：列表中所有文章的 `cover_image` 都是完整URL

## URL生成逻辑

### 优先级
1. **如果已经是完整URL**（以 `http://` 或 `https://` 开头）：直接返回
2. **从请求中获取**：
   - 检查 `X-Forwarded-Proto` 头（代理场景）
   - 检查请求的 TLS 状态（https）
   - 从请求头获取 `Host`
   - 拼接：`scheme://host/path`
3. **从配置获取**：使用 `configs.GetBaseURL()`
4. **返回相对路径**：如果以上都无法获取

### 示例

**请求**：`GET http://localhost:8080/api/admin/posts/1`

**返回**：
```json
{
  "post": {
    "id": 1,
    "title": "文章标题",
    "cover_image": "http://localhost:8080/uploads/images/20251103-xxx.jpg"  // ✅ 完整URL
  }
}
```

**之前**：
```json
{
  "post": {
    "id": 1,
    "title": "文章标题",
    "cover_image": "/uploads/images/20251103-xxx.jpg"  // ❌ 相对路径
  }
}
```

## 配置方式（可选）

如果需要手动设置基础URL（例如生产环境），可以在初始化时设置：

```go
import "api/configs"

func init() {
    // 设置基础URL
    configs.SetBaseURL("https://api.example.com")
}
```

如果不设置，系统会自动从请求中获取。

## 支持的场景

✅ **开发环境**：`http://localhost:8080`  
✅ **生产环境**：`https://api.example.com`  
✅ **代理/负载均衡器**：通过 `X-Forwarded-Proto` 头自动识别  
✅ **HTTPS**：自动检测 TLS 连接  
✅ **已有完整URL**：直接使用，不重复转换  

## 测试验证

### 1. 创建文章（带图片）
```bash
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer token" \
  -F "title=测试" \
  -F "content=内容" \
  -F "image=@/path/to/image.jpg"
```

**响应**：
```json
{
  "post": {
    "cover_image": "http://localhost:8080/uploads/images/20251103-xxx.jpg"
  }
}
```

### 2. 获取文章列表
```bash
curl http://localhost:8080/api/posts?page=1&size=10
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

## 修复总结

✅ **所有后台接口**：返回的图片URL都是完整的  
✅ **所有前端接口**：返回的图片URL都是完整的  
✅ **文件上传接口**：返回的URL都是完整的  
✅ **自动适配**：支持HTTP/HTTPS、代理场景  
✅ **向后兼容**：如果已经是完整URL，不会重复转换  

现在所有图片路径都可以直接访问了！🎉

