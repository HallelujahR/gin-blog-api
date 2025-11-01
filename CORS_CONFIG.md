# CORS跨域配置说明

## 问题背景

当后台请求接口时，如果携带了`Authorization`头部（token），浏览器会发起OPTIONS预检请求。如果服务器没有正确配置CORS，会出现跨域问题。

## 常见跨域错误

1. **OPTIONS请求被拒绝**：预检请求失败
2. **Authorization头部被阻止**：未配置允许Authorization头
3. **Credentials被拒绝**：AllowCredentials设置问题

## 解决方案

### 当前配置（开发环境）

在`middleware/cors.go`中，已经配置了完整的CORS支持：

```go
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowAllOrigins: true,              // 允许所有来源
		AllowMethods: [...]string{...},      // 允许所有方法
		AllowHeaders: [...]string{          // 重要：包含Authorization
			"Authorization",
			"Content-Type",
			...
		},
		AllowCredentials: true,             // 允许携带凭证
		MaxAge: 12 * time.Hour,            // 预检缓存12小时
	})
}
```

### 关键配置项说明

| 配置项 | 值 | 说明 |
|--------|-----|------|
| `AllowAllOrigins` | `true` | 允许所有来源访问（开发环境） |
| `AllowHeaders` | 包含`Authorization` | **关键**：允许携带token |
| `AllowCredentials` | `true` | 允许携带凭证 |
| `AllowMethods` | 全部HTTP方法 | 允许所有操作 |
| `MaxAge` | 12小时 | 预检请求缓存时间 |

## 生产环境配置

### 推荐配置

在生产环境中，应该限制允许的来源：

```go
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 生产环境：配置具体的允许来源
		AllowOrigins: []string{
			"https://yourdomain.com",
			"https://www.yourdomain.com",
			"https://admin.yourdomain.com",  // 后台管理域名
		},
		
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		
		AllowHeaders: []string{
			"Origin",
			"Content-Length",
			"Content-Type",
			"Authorization",  // 必须包含！
			"Accept",
			"X-Requested-With",
		},
		
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"Authorization",
		},
		
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
		AllowPrivateNetwork: true,
	})
}
```

### 环境变量配置（推荐）

可以通过环境变量动态配置：

```go
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders: []string{
			"Origin", "Content-Length", "Content-Type", 
			"Authorization", "Accept", "X-Requested-With",
		},
		ExposeHeaders: []string{"Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge: 12 * time.Hour,
		AllowPrivateNetwork: true,
	}
	
	// 根据环境变量配置允许的来源
	if os.Getenv("APP_ENV") == "production" {
		// 生产环境：从环境变量读取允许的域名
		origins := strings.Split(os.Getenv("ALLOWED_ORIGINS"), ",")
		config.AllowOrigins = origins
	} else {
		// 开发环境：允许所有来源
		config.AllowAllOrigins = true
	}
	
	return cors.New(config)
}
```

环境变量配置：
```bash
# .env文件
APP_ENV=production
ALLOWED_ORIGINS=https://yourdomain.com,https://admin.yourdomain.com
```

## 测试CORS配置

### 使用curl测试

```bash
# 测试预检请求（OPTIONS）
curl -X OPTIONS http://localhost:8080/api/admin/users \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Authorization" \
  -v

# 应该返回：
# Access-Control-Allow-Origin: *
# Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
# Access-Control-Allow-Headers: Authorization, Content-Type
# Access-Control-Allow-Credentials: true
```

### 使用浏览器控制台测试

```javascript
fetch('http://localhost:8080/api/admin/users', {
  method: 'GET',
  headers: {
    'Authorization': 'Bearer your-token-here',
    'Content-Type': 'application/json'
  },
  credentials: 'include'
})
.then(res => res.json())
.then(data => console.log(data))
.catch(err => console.error('CORS Error:', err));
```

## 常见问题排查

### 1. 仍然出现跨域错误

**检查清单**：
- ✅ 中间件是否在最前面加载？
- ✅ `AllowHeaders`中是否包含`Authorization`？
- ✅ 前端请求是否正确设置了`Authorization`头？
- ✅ 服务器是否重启？

### 2. OPTIONS请求被404

**原因**：路由没有处理OPTIONS方法
**解决**：确保`AllowMethods`中包含`OPTIONS`

### 3. 生产环境跨域失败

**原因**：`AllowOrigins`配置错误
**解决**：
- 确保域名格式正确（包含协议：`https://example.com`）
- 不要使用尾随斜杠
- 确保大小写一致

### 4. AllowCredentials问题

**规则**：
- ✅ 使用`AllowAllOrigins: true`时，不能同时用`AllowCredentials: true`
- ✅ 使用`AllowOrigins: []string{"*"}`时，不能同时用`AllowCredentials: true`
- ✅ 必须使用`AllowOrigins`指定具体域名 + `AllowCredentials: true`

## 安全建议

### 开发环境
- ✅ 使用`AllowAllOrigins: true`
- ⚠️ 仅在本地开发使用

### 生产环境
- ✅ 使用`AllowOrigins`指定具体域名
- ✅ 避免使用通配符`*`
- ✅ 限制允许的HTTP方法
- ✅ 限制允许的请求头
- ✅ 启用HTTPS
- ✅ 定期审查允许的域名列表

## 参考

- [MDN: CORS](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [gin-contrib/cors](https://github.com/gin-contrib/cors)
- [HTTP Preflight Request](https://developer.mozilla.org/en-US/docs/Glossary/Preflight_request)
