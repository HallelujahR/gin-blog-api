# 部署说明（无 Docker 版本）

本指南介绍如何在常规 Linux 服务器上，使用系统自带的 MySQL、Redis、Nginx 等服务部署本博客系统。所有步骤均基于直接安装的软件包，不再依赖 Docker 或容器脚本。

---

## 1. 系统要求
- 操作系统：Ubuntu 20.04+/Debian 11+/CentOS 7+/RockyLinux 8+
- Golang：1.25+
- Node.js：18+（仅当需要在服务器上构建前端时）
- MySQL：8.0.44（与开发环境一致）
- Redis：6+/7+
- Nginx：1.20+（或你熟悉的稳定版本）
- Git、Make、gcc、tzdata 等基础工具

> 建议以 `/www/wwwroot/blog/api` 作为后端目录，前端代码位于 `/www/wwwroot/blog/gin-blog-vue-font`。

---

## 2. 系统依赖安装

以 Ubuntu 为例：
```bash
sudo apt update
sudo apt install -y build-essential git curl nginx redis-server mysql-server
```

安装 Go 1.25：
```bash
wget https://go.dev/dl/go1.25.0.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.25.0.linux-amd64.tar.gz
cat <<'PROFILE' | sudo tee /etc/profile.d/go.sh
export PATH=$PATH:/usr/local/go/bin
PROFILE
source /etc/profile.d/go.sh
```

> 如果你已经通过其他方式安装，确保 `go version` 输出 1.25 及以上即可。

---

## 3. 数据库与 Redis 准备

### 3.1 初始化 MySQL
```bash
sudo systemctl enable --now mysql
mysql_secure_installation   # 根据提示设置 root 密码
```

创建业务库与账号（可按需调整密码）：
```bash
mysql -uroot -p <<'SQL'
CREATE DATABASE IF NOT EXISTS blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER IF NOT EXISTS 'blog_user'@'%' IDENTIFIED BY 'blog_password';
GRANT ALL PRIVILEGES ON blog.* TO 'blog_user'@'%';
FLUSH PRIVILEGES;
SQL
```

如需取消 `likes` 表的外键限制，可执行 `database/sql/fix_likes_foreign_key.sql`。

### 3.2 Redis
```bash
sudo systemctl enable --now redis-server
redis-cli ping   # 返回 PONG 代表正常
```

---

## 4. 拉取代码并配置环境
```bash
sudo mkdir -p /www/wwwroot/blog
cd /www/wwwroot/blog
git clone https://github.com/your-org/gin-blog-api.git api
cd api
cp env.template .env
vi .env
```

典型 `.env` 设置：
```
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=blog_password
DB_NAME=blog
BLOG_ENV=prod
API_BASE_URL=https://api.example.com
REDIS_ADDR=127.0.0.1:6379
```

> 建议将 `.env` 路径写入 systemd 的 `EnvironmentFile`，而不是导出到全局 profile。

同步依赖与构建：
```bash
go env -w GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
go mod download
go build -o bin/api ./
```

---

## 5. 运行后端服务

### 5.1 开发/调试
```bash
BLOG_ENV=dev go run main.go
```
日志写入 `logs/YYYY-MM-DD_log`，程序会自动创建目录与文件。

### 5.2 systemd 服务（推荐）
创建 `/etc/systemd/system/blog-api.service`：
```ini
[Unit]
Description=Blog API Service
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

启用：
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now blog-api
sudo systemctl status blog-api
```

---

## 6. 前端与 Nginx

前端代码位于 `/www/wwwroot/blog/gin-blog-vue-font`：
```bash
cd /www/wwwroot/blog/gin-blog-vue-font
npm ci
npm run build
```

Nginx 示例（`/etc/nginx/conf.d/blog.conf`）：
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

如需 HTTPS，为该 server block 配置证书（`listen 443 ssl;` + `ssl_certificate` 等）。

> 前端访问 API 时请使用相对路径 `/api`，避免写死 `localhost`。

---

## 7. 目录结构与日志
- `logs/YYYY-MM-DD_log`：Gin 访问日志，程序自动滚动
a- `uploads/`：图片、附件
- `database/sql/`：常用 SQL 脚本

日志轮转：可使用 `logrotate`，示例 `/etc/logrotate.d/blog-api`：
```
/www/wwwroot/blog/api/logs/*_log {
    daily
    rotate 30
    missingok
    compress
    notifempty
}
```

---

## 8. 更新与回滚
```bash
cd /www/wwwroot/blog/api
git pull
./bin/api --version   # 可自定义版本命令
# 重新编译
go build -o bin/api ./
sudo systemctl restart blog-api
```

回滚只需切换到旧 tag 并重新编译。

---

## 9. 故障排查

| 问题 | 排查要点 |
| --- | --- |
| API 无法启动 | `journalctl -u blog-api -xe`、确认 `.env` 加载、数据库可连接 |
| Redis 连接失败 | `redis-cli -h 127.0.0.1 ping`、检查 `REDIS_ADDR` |
| 前端 404 | 检查 Nginx `root` 和前端 `dist` 是否存在，`try_files` 是否正确 |
| `/api` 请求 404 | 确认 Nginx `proxy_pass http://127.0.0.1:8080;`，不要额外拼 `/api` 路径 |
| 日志缺失 | 确保运行用户对 `logs/` 目录有写权限 |

---

## 10. 附录
- `database/sql/init.sql`：示例用户/权限初始化
- `database/sql/fix_likes_foreign_key.sql`：移除 `likes` 外键
- 如果需要自动化部署，可自行编写 shell/systemd 单元，但不再提供 Docker 相关脚本

至此，服务器上的 Redis、MySQL、Nginx 均以系统服务运行，后端直接以二进制方式部署，满足“去 Docker 化”的要求。
