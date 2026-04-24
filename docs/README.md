# Blog API 文档

本目录只保留后端维护需要的文档，过细的接口示例、重复的项目介绍和历史性优化总结已合并删除。

## 文档索引

- [backend-architecture.md](backend-architecture.md)：目录结构、模块设计、现状评估与改进建议。
- [api-reference.md](api-reference.md)：公开接口、后台接口和认证约定的精简索引。
- [deployment.md](deployment.md)：本地启动、生产部署、环境变量与运维命令。
- [analytics-module.md](analytics-module.md)：访问统计链路说明。
- [draw-guess-realtime-sync.md](draw-guess-realtime-sync.md)：你画我猜实时同步设计。

## 快速启动

```bash
cp env.template .env
go mod download
go run ./cmd/server
```

默认监听 `:8080`，接口前缀为 `/api`，上传文件通过 `/uploads` 公开访问。

## 技术栈

- Web：Gin
- ORM：GORM
- 数据库：MySQL
- 缓存/统计：Redis
- 实时通信：Gorilla WebSocket
- 鉴权：数据库会话 token，格式为 `Authorization: Bearer <token>`
