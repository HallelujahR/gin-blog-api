# 后端结构与模块评估

## 当前目录结构

```text
api/
├── cmd/
│   ├── server/                 # API 服务入口与启动装配
│   └── tools/                  # 日志转换、回放、图片优化等运维工具
├── internal/
│   ├── config/                 # 环境变量配置
│   ├── middleware/             # CORS、鉴权、限流、访问采集
│   ├── platform/               # DB、Redis、日志、GeoIP 等基础设施
│   │   ├── db/
│   │   ├── geoip/
│   │   ├── logstore/
│   │   └── redisstore/
│   └── modules/
│       ├── content/            # 博客内容域：路由、控制器、业务、DAO、模型
│       ├── analytics/          # 访问事件采集、ETL、统计 API
│       ├── media/              # 上传、图片压缩、本地图片优化
│       └── drawguess/          # 你画我猜房间状态与 WebSocket
├── database/
│   └── sql/                    # 初始化与变更 SQL
├── docs/                       # 项目文档
├── scripts/                    # 部署/批处理脚本
└── uploads/                    # 上传文件运行时目录，已被 .gitignore 忽略
```

## 启动链路

`cmd/server/main.go` 调用 `Run()`，执行顺序为：

1. `InitConfig()`：加载环境变量配置。
2. `InitDB()`：连接 MySQL，按模型确保表存在。
3. 可选启动 `analytics.StartETLWorker()`。
4. `InitRouter()`：注册中间件、静态文件和公开/后台路由。
5. `r.Run(:HTTP_PORT)`。

## 结构评价

当前已经从“根目录横向分层”整理成“入口 + internal + modules”的结构，根目录明显干净了。`internal/platform` 承担基础设施，`internal/middleware` 承担请求链路能力，业务代码收在 `internal/modules` 下。

`content` 目前仍是过渡模块，里面保留了原来的 `routes/controllers/service/dao/models` 分层，覆盖文章、分类、标签、页面、动态、评论、用户、留言、点赞、热点数据等博客主业务。这样做能减少一次性重构风险；后续如果继续拆，可以再从 `content` 内部分出 `comment`、`user` 等子模块。

## 合理之处

- 单体服务选择正确：博客、管理后台、统计、小游戏目前共享数据库与部署生命周期，拆微服务成本高于收益。
- `cmd/server` 只负责启动服务，`cmd/tools` 与主服务入口隔离。
- `internal/platform` 把 DB、Redis、日志、GeoIP 从业务服务中拿出来，边界更清楚。
- `analytics`、`media`、`drawguess` 已按模块聚合，阅读路径比原来短。
- 运行目录 `uploads/`、`logs/`、`data/` 不再作为代码目录参与版本管理。

## 仍需改进

- `internal/modules/content` 仍偏大，可继续拆出 `user`、`comment` 或 `taxonomy`。
- 公开路由中部分写接口仍暴露，如 `POST /api/posts`、`POST /api/categories`、`PUT /api/pages/:id`。如果前台不需要写，应移动到后台或加鉴权。
- `InitDB()` 只在表不存在时 `AutoMigrate`，已有表字段变更不会自动迁移。生产变更建议使用明确 SQL/migration。
- DAO 和平台 DB 仍通过全局单例调用，简单但不利于测试。后续可在模块边界引入依赖注入。

## 推荐演进方向

1. 收紧公开写接口权限，明确哪些是前台匿名接口，哪些只能后台调用。
2. 从 `internal/modules/content` 内继续拆出 `user` 和 `comment`，但每次只迁一个业务域。
3. 数据库结构变更从 `AutoMigrate` 迁到可审计 SQL 脚本，保留 `AutoMigrate` 只用于本地开发或首次建表。
