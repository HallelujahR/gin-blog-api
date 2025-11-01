# Blog API 架构说明

## 架构概述

本项目采用**单服务双模块**架构，将前端用户访问API和后台管理API进行了明确的分离，同时最大化复用现有的代码（dao、service层）。

## 为什么选择单服务而不是微服务？

对于博客系统这种规模的应用：
1. **开发效率高**：减少服务间通信复杂度，统一依赖管理
2. **资源占用少**：单进程运行，内存和CPU占用更低
3. **部署简单**：只需部署一个服务，无需复杂的服务治理
4. **调试方便**：本地调试无需多服务启动，问题排查更容易
5. **数据一致性**：共享同一个数据库连接，避免分布式事务
6. **代码复用**：dao和service层完全复用，减少代码重复

如果未来业务增长到需要单独扩展某些模块（如搜索、推荐等），可以再将特定模块独立为微服务。

## 架构分层

### 1. 路由层（Routes）

#### 前端用户API (`routes/`)
- 路径前缀：`/api/xxx`
- 特点：
  - **完全兼容现有API**，保持原有接口地址不变
  - 支持匿名访问（如评论、点赞）
  - 支持登录用户访问（用户信息更新等）
  
#### 后台管理API (`routes/admin/`)
- 路径前缀：`/api/admin/xxx`
- 特点：
  - **严格的权限控制**：需要管理员身份
  - 统一的认证中间件：`AuthMiddleware` + `AdminMiddleware`
  - 提供管理功能：用户管理、内容审核、系统配置等

### 2. 控制器层（Controllers）

#### 前端控制器 (`controllers/`)
- `user_controller.go` - 用户注册、登录、个人资料
- `post_controller.go` - 文章列表、详情（公开访问）
- `comment_controller.go` - 评论创建、列表（支持匿名）
- `like_controller.go` - 点赞（支持匿名）
- 等其他控制器...

#### 管理控制器 (`controllers/admin/`)
- `user_controller.go` - 用户列表、删除、状态管理
- `post_controller.go` - 文章创建、编辑、删除
- `category_controller.go` - 分类管理
- `tag_controller.go` - 标签管理
- `comment_controller.go` - 评论审核、删除

### 3. 业务逻辑层（Services）

**完全复用**，前后端使用相同的service函数：
- `service/user_service.go`
- `service/post_service.go`
- `service/category_service.go`
- `service/tag_service.go`
- `service/comment_service.go`

### 4. 数据访问层（DAOs）

**完全复用**，前后端使用相同的dao函数：
- `dao/user_dao.go`
- `dao/post_dao.go`
- `dao/category_dao.go`
- `dao/tag_dao.go`
- `dao/comment_dao.go`
- `dao/session_dao.go` - 新增：会话管理

### 5. 中间件层（Middleware）

新建 `middleware/auth.go`：

#### AuthMiddleware - 用户认证
- 验证Bearer Token
- 检查会话是否过期
- 验证用户状态是否激活
- 将用户信息存储到上下文（c.Set）

#### AdminMiddleware - 管理员权限
- 验证用户角色是否为`admin`
- 与AuthMiddleware配合使用

#### AuthorOrAdminMiddleware - 作者或管理员
- 验证用户角色为`author`或`admin`

#### OptionalAuthMiddleware - 可选认证
- 支持匿名访问
- 如果有有效token则填充用户信息到上下文

### 6. 模型层（Models）

**完全复用**：
- `models/user.go` - 用户模型
- `models/user_session.go` - 会话模型
- `models/post.go` - 文章模型
- `models/comment.go` - 评论模型
- 等其他模型...

## 认证流程

### 登录流程
```
1. POST /api/users/login
   {
     "username": "admin",
     "password": "123456"
   }
   
2. 返回：
   {
     "user": {...},
     "token": "uuid1_uuid2"
   }
   
3. 前端存储token
```

### 调用需要认证的API
```
Authorization: Bearer {token}
```

### Token管理
- 生成：登录成功后使用`middleware.CreateSession()`创建会话
- 验证：`AuthMiddleware`自动验证
- 过期：7天后自动失效
- 刷新：重新登录获取新token
- 登出：前端删除token即可（或调用登出接口删除会话）

## API端点列表

### 前端用户API

| 功能 | 方法 | 路径 | 权限 |
|------|------|------|------|
| 用户注册 | POST | `/api/users` | 公开 |
| 用户登录 | POST | `/api/users/login` | 公开 |
| 获取用户信息 | GET | `/api/users/:id` | 公开 |
| 更新用户资料 | PUT | `/api/users/:id` | 可选认证 |
| 删除用户 | DELETE | `/api/users/:id` | 认证 |
| 文章列表 | GET | `/api/posts` | 公开 |
| 文章详情 | GET | `/api/posts/:id` | 公开 |
| 创建评论 | POST | `/api/comments` | 公开（匿名） |
| 评论列表 | GET | `/api/comments?post_id=xxx` | 公开 |
| 点赞/取消 | POST | `/api/like/toggle` | 公开（匿名） |
| 点赞统计 | GET | `/api/like/count` | 公开 |

### 后台管理API

所有管理API都需要管理员权限！

| 功能 | 方法 | 路径 | 权限 |
|------|------|------|------|
| 用户列表 | GET | `/api/admin/users` | 管理员 |
| 删除用户 | DELETE | `/api/admin/users/:id` | 管理员 |
| 更新用户状态 | PUT | `/api/admin/users/:id/status` | 管理员 |
| 更新用户角色 | PUT | `/api/admin/users/:id/role` | 管理员 |
| 创建文章 | POST | `/api/admin/posts` | 管理员 |
| 更新文章 | PUT | `/api/admin/posts/:id` | 管理员 |
| 删除文章 | DELETE | `/api/admin/posts/:id` | 管理员 |
| 创建分类 | POST | `/api/admin/categories` | 管理员 |
| 更新分类 | PUT | `/api/admin/categories/:id` | 管理员 |
| 删除分类 | DELETE | `/api/admin/categories/:id` | 管理员 |
| 创建标签 | POST | `/api/admin/tags` | 管理员 |
| 更新标签 | PUT | `/api/admin/tags/:id` | 管理员 |
| 删除标签 | DELETE | `/api/admin/tags/:id` | 管理员 |
| 评论列表 | GET | `/api/admin/comments` | 管理员 |
| 删除评论 | DELETE | `/api/admin/comments/:id` | 管理员 |
| 更新评论状态 | PUT | `/api/admin/comments/:id/status` | 管理员 |

## 前端是否需要登录功能？

### 当前设计（不需要登录）
- **评论**：使用`author_name`和`author_email`匿名评论
- **点赞**：使用`user_id`标识，可以是匿名用户ID（前端生成）

### 可选增强（支持登录）
如果未来需要支持用户登录，可以：
1. 评论时自动填充已登录用户信息
2. 点赞时自动使用当前登录用户ID
3. 用户的个人中心功能

**当前架构已完全支持这两种模式，通过`OptionalAuthMiddleware`实现**

## 数据库表结构

### 新增表：user_sessions
```sql
CREATE TABLE user_sessions (
  id BIGINT PRIMARY KEY AUTO_INCREMENT,
  user_id BIGINT NOT NULL,
  session_token VARCHAR(255) UNIQUE NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  user_agent TEXT,
  ip_address VARCHAR(45),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  INDEX idx_user_id (user_id),
  INDEX idx_token (session_token)
);
```

## 安全性

### 已实现的安全措施
1. **密码加密**：bcrypt哈希
2. **Token机制**：UUID生成的唯一token
3. **会话过期**：7天自动失效
4. **权限控制**：基于角色的访问控制（RBAC）
5. **状态验证**：检查用户是否被禁用

### 建议添加的安全措施（可选）
1. 刷新Token机制
2. Token黑名单（登出时标记失效）
3. 登录频率限制（防止暴力破解）
4. IP白名单（管理后台）
5. CORS白名单配置

## 部署建议

### 开发环境
```bash
go run main.go
```

### 生产环境
1. 使用supervisor或systemd管理进程
2. 配置反向代理（nginx/caddy）
3. 配置HTTPS
4. 环境变量管理敏感信息
5. 日志采集和监控

## 扩展性

### 未来可能的扩展
1. **缓存层**：Redis缓存热点数据
2. **搜索引擎**：Elasticsearch全文搜索
3. **CDN**：静态资源CDN加速
4. **消息队列**：异步任务处理
5. **监控系统**：APM和日志分析

### 迁移到微服务的时机
当出现以下情况时考虑拆分：
- 搜索功能需要独立扩展
- 推荐算法需要独立部署
- 某些模块需要不同的技术栈
- 团队规模扩大，需要独立开发

## 总结

本架构的设计原则：
1. ✅ **向前兼容**：保持原有API地址不变
2. ✅ **代码复用**：dao和service层完全共享
3. ✅ **权限分离**：前端公开访问，后台严格管控
4. ✅ **易于维护**：清晰的目录结构
5. ✅ **扩展友好**：预留了多种扩展空间

## 联系与反馈

如有问题或建议，请提交Issue或PR。
