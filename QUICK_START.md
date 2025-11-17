# 快速启动（无 Docker）

本文档介绍如何在直接安装的 MySQL、Redis、Nginx 环境中部署 Golang 博客 API。

## 1. 系统要求
- Linux x86_64（建议 Ubuntu 20.04+/CentOS 7+/RockyLinux 8+）
- Go 1.25+
- Node.js 18+（仅在需要构建前端时）
- MySQL 8.0.44
- Redis 6+/7+
- Nginx（用于反向代理和前端静态资源）

## 2. 准备数据库与 Redis
```bash
sudo systemctl enable --now mysqld   # 或 mariadb
sudo systemctl enable --now redis
```

初始化数据库账号（可根据需要修改密码）：
```bash
mysql -uroot -p <<'SQL'
CREATE DATABASE IF NOT EXISTS blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'blog_user'@'%' IDENTIFIED BY 'blog_password';
GRANT ALL ON blog.* TO 'blog_user'@'%';
FLUSH PRIVILEGES;
SQL
```

如果需要删除 `likes` 表的外键，可执行 `database/sql/fix_likes_foreign_key.sql`。

## 3. 配置环境变量
```bash
cp env.template .env
vi .env
```
关键变量：
```
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=blog_password
DB_NAME=blog
BLOG_ENV=prod
API_BASE_URL=https://your-domain.com
REDIS_ADDR=127.0.0.1:6379
```

加载 `.env`（可写入 `/etc/profile.d/blog.sh`）：
```bash
export $(grep -v '^#' .env | xargs)
```

## 4. 启动后端 API
```bash
go mod download
go build -o bin/api ./
./bin/api
```

推荐使用 systemd 管理：
```ini
# /etc/systemd/system/blog-api.service
[Unit]
Description=Blog API (Go)
After=network.target mysql.service redis.service

[Service]
Type=simple
WorkingDirectory=/www/wwwroot/blog/api
EnvironmentFile=/www/wwwroot/blog/api/.env
ExecStart=/www/wwwroot/blog/api/bin/api
Restart=on-failure

[Install]
WantedBy=multi-user.target
```
启用：
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now blog-api
```

## 5. 前端与 Nginx（示例）
1. 在 `/www/wwwroot/blog/gin-blog-vue-font` 拉取前端代码并执行 `npm ci && npm run build`。
2. 配置 Nginx：
```nginx
server {
    listen 80;
    server_name example.com;
    root /www/wwwroot/blog/gin-blog-vue-font/dist;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:8080/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```
3. 重新加载 Nginx：`sudo nginx -s reload`。

## 6. 常见问题
- **端口占用**：`ss -tlnp | grep 8080` 并停止冲突进程。
- **数据库连接失败**：检查 `.env` 是否加载，`mysql -h127.0.0.1 -ublog_user -p` 能否登录。
- **Redis 连接失败**：确认 `redis-cli ping` 返回 `PONG`。
- **日志位置**：所有访问日志写入 `logs/YYYY-MM-DD_log`，启动程序即可自动创建，无需手动 `touch`。

## 7. 更新代码
```bash
cd /www/wwwroot/blog/api
git pull
go build -o bin/api ./
sudo systemctl restart blog-api
sudo systemctl reload nginx
```
