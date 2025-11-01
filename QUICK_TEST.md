# 快速测试指南

## 1. 启动服务

```bash
cd /Users/wangruiwen/Desktop/blog/api
go run main.go
```

服务启动在：http://localhost:8080

## 2. 测试CORS配置

### 测试OPTIONS预检请求

```bash
curl -X OPTIONS http://localhost:8080/api/admin/users \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Authorization" \
  -v 2>&1 | grep -i "access-control"
```

**期望返回**：
```
< access-control-allow-origin: *
< access-control-allow-methods: GET, POST, PUT, PATCH, DELETE, OPTIONS, HEAD
< access-control-allow-headers: Origin, Content-Length, Content-Type, Authorization, Accept, X-Requested-With
< access-control-allow-credentials: true
```

### 测试前端请求（浏览器控制台）

```javascript
// 1. 先登录获取token
fetch('http://localhost:8080/api/users/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify({
    username: 'admin',
    password: '123456'
  })
})
.then(res => res.json())
.then(data => {
  console.log('Login success:', data);
  const token = data.token;
  
  // 2. 使用token调用管理接口
  return fetch('http://localhost:8080/api/admin/users', {
    method: 'GET',
    headers: {
      'Authorization': 'Bearer ' + token,
      'Content-Type': 'application/json'
    }
  });
})
.then(res => res.json())
.then(data => console.log('Admin API success:', data))
.catch(err => console.error('Error:', err));
```

## 3. 测试原有接口兼容性

### 测试文章列表（公开API）

```bash
curl http://localhost:8080/api/posts?page=1&size=10
```

### 测试评论创建（支持匿名）

```bash
curl -X POST http://localhost:8080/api/comments \
  -H "Content-Type: application/json" \
  -d '{
    "content": "测试评论",
    "author_name": "测试用户",
    "author_email": "test@example.com",
    "post_id": 1
  }'
```

### 测试点赞（支持匿名）

```bash
curl -X POST http://localhost:8080/api/like/toggle \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "post_id": 1
  }'
```

## 4. 测试新的管理接口

### 先创建admin用户（如果需要）

```sql
-- 在数据库中执行
INSERT INTO users (username, email, password_hash, role, status) 
VALUES ('admin', 'admin@example.com', '$2a$10$....', 'admin', 'active');
```

### 测试登录（获取token）

```bash
curl -X POST http://localhost:8080/api/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "123456"
  }'
```

### 使用token测试管理接口

```bash
# 替换YOUR_TOKEN为实际返回的token
TOKEN="your-actual-token-here"

# 获取用户列表
curl http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer $TOKEN"

# 创建文章
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试文章",
    "slug": "test-post",
    "content": "这是测试内容"
  }'

# 创建分类
curl -X POST http://localhost:8080/api/admin/categories \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试分类",
    "slug": "test-category"
  }'
```

## 5. 测试权限控制

### 测试未登录访问（应返回401）

```bash
curl -X GET http://localhost:8080/api/admin/users \
  -H "Content-Type: application/json"
```

**期望返回**：
```json
{"error": "未提供认证信息"}
```

### 测试无效token（应返回401）

```bash
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer invalid-token" \
  -H "Content-Type: application/json"
```

**期望返回**：
```json
{"error": "无效的认证令牌"}
```

## 6. 常见问题排查

### 问题1：跨域错误仍然存在

**排查步骤**：
1. 确认服务已重启
2. 检查浏览器Network面板，查看OPTIONS请求的响应头
3. 确认`Access-Control-Allow-Headers`包含`Authorization`

### 问题2：登录接口不返回token

**排查步骤**：
1. 检查数据库连接
2. 确认user_sessions表已创建
3. 查看服务日志

### 问题3：管理接口403错误

**排查步骤**：
1. 确认用户role为`admin`
2. 确认Token有效
3. 检查中间件加载顺序

## 7. 数据库初始化

如果需要创建user_sessions表：

```sql
CREATE TABLE IF NOT EXISTS user_sessions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  session_token VARCHAR(255) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  user_agent TEXT,
  ip_address VARCHAR(45),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_token (session_token)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

## 8. 生产环境配置

### 修改CORS配置

编辑`middleware/cors.go`：

```go
func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		// 生产环境：改为具体域名
		AllowOrigins: []string{
			"https://yourdomain.com",
			"https://admin.yourdomain.com",
		},
		// ... 其他配置保持不变
	})
}
```

### 设置环境变量

```bash
export BLOG_ENV=prod
export DB_HOST=your-prod-db-host
export DB_USER=your-prod-db-user
export DB_PASSWORD=your-prod-db-password
export DB_NAME=your-prod-db-name
```

## 9. 性能优化建议

1. **缓存**：考虑添加Redis缓存
2. **限流**：添加请求限流中间件
3. **日志**：配置结构化日志
4. **监控**：添加APM监控

## 10. 安全加固建议

1. **HTTPS**：生产环境使用HTTPS
2. **Token刷新**：实现refresh token机制
3. **IP白名单**：限制管理后台IP访问
4. **日志审计**：记录所有管理操作
5. **SQL注入防护**：确保使用参数化查询（已使用GORM）

## 完成！

现在你的API已经支持：
- ✅ CORS跨域（包含Authorization支持）
- ✅ Token认证
- ✅ 权限控制
- ✅ 前后端API分离
- ✅ 所有原有接口兼容

祝你使用愉快！🎉
