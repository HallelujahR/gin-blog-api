# 部署与运维

## 环境要求

- Go 1.25+
- MySQL 8 或兼容版本
- Redis 6/7，可关闭但统计缓存与访问分析会降级
- Nginx，用于前端静态资源、`/api` 反代和 WebSocket Upgrade

## 环境变量

复制模板：

```bash
cp env.template .env
```

常用变量：

```env
BLOG_ENV=prod
HTTP_PORT=8080
API_BASE_URL=https://example.com

DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=blog
DB_PASSWORD=blog_password
DB_NAME=blog
DB_MAX_OPEN_CONNS=8
DB_MAX_IDLE_CONNS=3
DB_LOG_LEVEL=warn

ENABLE_REDIS=true
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0

ENABLE_ACCESS_LOG=true
ENABLE_ANALYTICS=false
ANALYTICS_ETL_INTERVAL=30m
ENABLE_PPROF=false
PPROF_PORT=6060
```

## 本地运行

```bash
go mod download
go run ./cmd/server
```

服务默认监听 `http://localhost:8080`。

## 生产构建

```bash
go build -o bin/api ./cmd/server
./bin/api
```

推荐用 systemd 托管：

```ini
[Unit]
Description=Blog API
After=network.target mysql.service redis.service

[Service]
Type=simple
WorkingDirectory=/www/wwwroot/blog/api
EnvironmentFile=/www/wwwroot/blog/api/.env
ExecStart=/www/wwwroot/blog/api/bin/api
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

## Nginx 反代

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_read_timeout 600s;
}
```

`/uploads` 由后端静态服务提供；如果改由 Nginx 直接托管，需要保证目录权限和 URL 前缀一致。

## 数据库

首次启动会自动创建缺失表。生产环境字段变更建议使用 `database/sql/` 中的 SQL 脚本或正式 migration，而不是依赖自动迁移。

常用脚本：

- `database/sql/init.sql`：初始化数据。
- `database/sql/performance_indexes.sql`：补充常用查询索引。
- `database/sql/fix_likes_foreign_key.sql`：修复历史点赞外键问题。

## 运维命令

```bash
# 重建并重启
go build -o bin/api ./cmd/server
sudo systemctl restart blog-api

# 查看日志
journalctl -u blog-api -f

# 回放访问日志到 Redis
go run ./cmd/tools/replay_raw_logs -log-dir=./logs -start=YYYY-MM-DD -end=YYYY-MM-DD

# 优化上传图片
go run ./cmd/tools/optimize_uploaded_images
```
