# Blog API 后端服务

基于 Gin 框架的博客系统后端 API 服务。

## 项目简介

这是一个功能完整的博客系统后端 API，支持文章管理、分类标签、评论点赞、用户管理、页面管理等核心功能。采用前后端分离架构，提供 RESTful API 接口。

## 技术栈

- **框架**: Gin (Go Web Framework)
- **数据库**: MySQL
- **ORM**: GORM
- **认证**: Token-based Authentication
- **跨域**: CORS Middleware

## 项目结构

```
api/
├── configs/          # 配置管理
├── controllers/      # 控制器层（处理HTTP请求）
│   └── admin/       # 后台管理控制器
├── dao/             # 数据访问层（数据库操作）
├── database/        # 数据库连接
├── initializer/     # 初始化模块
├── middleware/      # 中间件（认证、CORS等）
├── models/          # 数据模型
├── routes/          # 路由配置
│   └── admin/      # 后台管理路由
├── service/         # 业务逻辑层
├── uploads/         # 文件上传目录
└── main.go          # 入口文件
```

## 快速开始

### 环境要求

- Go 1.19+
- MySQL 5.7+

### 安装依赖

```bash
go mod download
```

### 配置数据库

修改数据库连接配置（在 `database/db.go` 或配置文件中）。

### 运行服务

```bash
go run main.go
```

服务默认运行在 `http://localhost:8080`

## 主要功能

### 前端用户接口

- ✅ 文章列表和详情（支持分页、筛选、搜索）
- ✅ 分类和标签管理
- ✅ 评论系统（支持嵌套回复）
- ✅ 点赞功能
- ✅ 热点数据展示
- ✅ 关于我页面

### 后台管理接口

- ✅ 文章管理（创建、编辑、删除、发布）
- ✅ 分类和标签管理
- ✅ 评论管理（审核、删除、回复）
- ✅ 用户管理
- ✅ 页面管理（富文本编辑器）
- ✅ 文件上传（图片、文件）

## API 文档

详细的接口文档请参考：[API_DOCUMENTATION.md](./API_DOCUMENTATION.md)

## 配置说明

### 基础URL配置

在 `configs/config.go` 中配置 API 服务的基础URL：

```go
// 默认开发环境
BaseURL = "http://localhost:8080"

// 生产环境可通过环境变量或代码设置
configs.SetBaseURL("https://api.example.com")
```

## 文件上传

上传的文件存储在 `uploads/` 目录：
- `uploads/images/`: 图片文件
- `uploads/files/`: 其他文件

上传的文件可通过 `/uploads/` 路径公开访问。

## 认证授权

### Token认证

大部分接口需要 Bearer Token 认证：

```
Authorization: Bearer {token}
```

### 角色权限

- **普通用户**: 可访问前端接口
- **管理员**: 可访问所有后台管理接口

## 开发说明

### 代码架构

采用 MVC 分层架构：
- **Controller**: 处理HTTP请求和响应
- **Service**: 业务逻辑处理
- **DAO**: 数据访问抽象层
- **Model**: 数据模型定义

### 路由分组

- `/api/`: 前端用户接口
- `/api/admin/`: 后台管理接口（需要管理员权限）

## License

MIT License

