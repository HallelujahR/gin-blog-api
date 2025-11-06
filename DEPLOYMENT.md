# 部署文档

## 服务器要求

### 最低配置

- CPU: 2 核
- 内存: 2GB
- 硬盘: 20GB
- 操作系统: Ubuntu 20.04+ / CentOS 7+ / Debian 10+

### 必需软件

- Docker 20.10+
- Docker Compose 2.0+
- Git

## 部署步骤

### 1. 服务器准备

#### 安装 Docker

**Ubuntu/Debian:**

```bash
# 更新系统
sudo apt update && sudo apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 安装Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 将当前用户添加到docker组（可选，避免每次使用sudo）
sudo usermod -aG docker $USER
```

**CentOS:**

```bash
# 安装Docker
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker
sudo systemctl start docker
sudo systemctl enable docker
```

#### 验证安装

```bash
docker --version
docker-compose --version
```

### 2. 克隆代码

```bash
# 创建项目目录
sudo mkdir -p /opt/blog
sudo chown $USER:$USER /opt/blog
cd /opt/blog

# 克隆后端代码
git clone https://github.com/your-username/gin-blog-api.git api
cd api

# 克隆前端代码（如果前端在单独仓库）
cd ..
git clone https://github.com/your-username/blog-front.git front
```

### 3. 配置环境变量

```bash
cd /opt/blog/api

# 复制环境变量模板
cp .env.example .env

# 编辑环境变量
nano .env
```

**配置内容**:

```env
# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=your_strong_password_here
DB_NAME=blog

# MySQL Root密码
MYSQL_ROOT_PASSWORD=your_root_password_here

# API基础URL（替换为你的域名）
API_BASE_URL=https://api.yourdomain.com

# 环境标识
BLOG_ENV=prod
```

### 4. 修改数据库配置

编辑 `database/db.go`，更新生产环境数据库连接：

```go
func getDSN() string {
    env := os.Getenv("BLOG_ENV")
    if env == "prod" {
        // 从环境变量读取数据库配置
        dbHost := os.Getenv("DB_HOST")
        dbPort := os.Getenv("DB_PORT")
        dbUser := os.Getenv("DB_USER")
        dbPassword := os.Getenv("DB_PASSWORD")
        dbName := os.Getenv("DB_NAME")

        if dbPort == "" {
            dbPort = "3306"
        }

        return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            dbUser, dbPassword, dbHost, dbPort, dbName)
    }
    // 开发环境配置...
    return "root:10244201@tcp(127.0.0.1:3306)/blog?charset=utf8mb4&parseTime=True&loc=Local"
}
```

### 5. 配置前端 API 地址

编辑 `/opt/blog/front/src/api/index.js`，修改 API 基础地址：

```javascript
const http = axios.create({
  baseURL: "https://api.yourdomain.com/api", // 修改为你的API地址
  timeout: 7000,
});
```

### 6. 首次部署

```bash
cd /opt/blog/api

# 给部署脚本执行权限
chmod +x scripts/*.sh

# 运行部署脚本
./scripts/deploy.sh production
```

### 7. 验证部署

```bash
# 查看容器状态
docker-compose -f docker-compose.prod.yml ps

# 查看日志
docker-compose -f docker-compose.prod.yml logs -f

# 测试API
curl http://localhost:8080/api/posts?page=1&size=1

# 测试前端
curl http://localhost
```

## 自动化部署（GitHub Actions）

### 1. 配置 GitHub Secrets

在 GitHub 仓库设置中添加以下 Secrets：

- `SERVER_HOST`: 服务器 IP 地址
- `SERVER_USER`: SSH 用户名
- `SERVER_SSH_KEY`: SSH 私钥
- `SERVER_PORT`: SSH 端口（默认 22）

### 2. SSH 密钥配置

在服务器上生成 SSH 密钥对（如果还没有）：

```bash
# 在本地生成密钥
ssh-keygen -t rsa -b 4096 -C "github-actions"

# 将公钥添加到服务器的authorized_keys
cat ~/.ssh/id_rsa.pub | ssh user@your-server "mkdir -p ~/.ssh && cat >> ~/.ssh/authorized_keys"
```

### 3. 推送代码触发部署

```bash
git add .
git commit -m "部署配置"
git push origin main
```

推送到 main 分支后，GitHub Actions 会自动部署到服务器。

## 常用操作

### 查看日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f api
docker-compose -f docker-compose.prod.yml logs -f frontend
docker-compose -f docker-compose.prod.yml logs -f mysql
```

### 重启服务

```bash
# 重启所有服务
docker-compose -f docker-compose.prod.yml restart

# 重启特定服务
docker-compose -f docker-compose.prod.yml restart api
```

### 更新代码

```bash
# 方式1：使用更新脚本（推荐，零停机）
./scripts/update.sh production

# 方式2：重新部署
./scripts/deploy.sh production
```

### 停止服务

```bash
docker-compose -f docker-compose.prod.yml down
```

### 进入容器

```bash
# 进入API容器
docker exec -it blog-api sh

# 进入MySQL容器
docker exec -it blog-mysql mysql -u blog_user -p
```

### 备份数据库

```bash
# 创建备份脚本
cat > scripts/backup-db.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/opt/blog/backups"
DATE=$(date +%Y%m%d_%H%M%S)
mkdir -p $BACKUP_DIR

docker exec blog-mysql mysqldump -u blog_user -p$DB_PASSWORD blog > $BACKUP_DIR/backup_$DATE.sql
echo "✅ 备份完成: $BACKUP_DIR/backup_$DATE.sql"
EOF

chmod +x scripts/backup-db.sh
```

## 域名和 HTTPS 配置

### 使用 Nginx 反向代理（推荐）

如果使用外部 Nginx，配置如下：

```nginx
# /etc/nginx/sites-available/blog
server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    # 重定向到HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name yourdomain.com www.yourdomain.com;

    ssl_certificate /etc/letsencrypt/live/yourdomain.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/yourdomain.com/privkey.pem;

    # 前端
    location / {
        proxy_pass http://localhost:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API（如果API使用独立域名）
    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 使用 Let's Encrypt SSL 证书

```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d yourdomain.com -d www.yourdomain.com
```

## 监控和维护

### 健康检查

```bash
# API健康检查
curl http://localhost:8080/api/posts?page=1&size=1

# 前端健康检查
curl http://localhost/health
```

### 资源监控

```bash
# 查看容器资源使用
docker stats

# 查看磁盘使用
df -h
docker system df
```

### 日志轮转

创建日志轮转配置：

```bash
cat > /etc/logrotate.d/docker-containers << 'EOF'
/var/lib/docker/containers/*/*.log {
    rotate 7
    daily
    compress
    size=1M
    missingok
    delaycompress
    copytruncate
}
EOF
```

## 故障排查

### 服务无法启动

```bash
# 查看详细日志
docker-compose -f docker-compose.prod.yml logs

# 检查端口占用
netstat -tulpn | grep -E '8080|80|3306'

# 检查磁盘空间
df -h
```

### 数据库连接失败

```bash
# 检查MySQL容器状态
docker ps | grep mysql

# 检查MySQL日志
docker logs blog-mysql

# 测试数据库连接
docker exec -it blog-mysql mysql -u blog_user -p
```

### 前端无法访问 API

```bash
# 检查API服务状态
docker ps | grep api

# 检查API日志
docker logs blog-api

# 测试API连接
curl http://localhost:8080/api/posts?page=1&size=1
```

## 安全建议

1. **修改默认密码**: 确保所有默认密码都已修改
2. **防火墙配置**: 只开放必要端口（80, 443, 22）
3. **定期备份**: 设置数据库自动备份
4. **更新系统**: 定期更新系统和 Docker 镜像
5. **日志监控**: 设置日志监控和告警
6. **SSL 证书**: 使用 HTTPS 加密传输

## 性能优化

1. **数据库索引**: 确保数据库表有适当的索引
2. **缓存策略**: 考虑使用 Redis 缓存热点数据
3. **CDN 加速**: 静态资源使用 CDN 加速
4. **负载均衡**: 高并发场景考虑使用负载均衡

## 快速开始（CentOS 服务器）

### CentOS 服务器一键部署步骤

```bash
# 1. 安装Docker（CentOS 7/8/9）
# CentOS 7
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# CentOS 8/9 或 Rocky Linux
sudo dnf install -y yum-utils
sudo dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker服务
sudo systemctl start docker && sudo systemctl enable docker

# 验证安装
docker --version
docker compose version

# 2. 配置防火墙（开放必要端口）
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --reload

# 3. 克隆代码
cd /opt && sudo mkdir -p blog && sudo chown $USER:$USER blog
cd blog
git clone https://github.com/HallelujahR/gin-blog-api.git api
cd api

# 4. 配置环境变量
cp env.template .env
vi .env  # 编辑数据库和API地址（必须修改密码）

# 5. 配置镜像加速器（强烈推荐，解决拉取超时）
chmod +x scripts/*.sh
sudo ./scripts/configure-docker-mirror.sh

# 6. 部署
./scripts/deploy.sh production

# 7. 验证部署
docker compose -f docker-compose.prod.yml ps
curl http://localhost:8080/api/posts?page=1&size=1
```

## 部署检查清单

### 部署前准备

- [ ] 服务器已安装 Docker 和 Docker Compose
- [ ] 服务器已开放必要端口（80, 443, 8080, 22）
- [ ] 服务器有足够的资源（CPU、内存、磁盘）
- [ ] 代码已推送到 GitHub 仓库
- [ ] 已配置数据库连接信息
- [ ] 已配置 API 基础 URL

### 部署后验证

- [ ] Docker 和 Docker Compose 已正确安装
- [ ] 所有容器都在运行（3 个容器）
- [ ] API 接口可以正常访问
- [ ] 前端页面可以正常访问
- [ ] 数据库连接正常
- [ ] 文件上传功能正常
- [ ] 日志无错误信息

## 常见问题排查

### Docker 服务启动失败

**错误信息：**

```
Job for docker.service failed because the control process exited with error code.
```

**解决方案：**

```bash
# 使用修复脚本
sudo ./scripts/fix-docker-service.sh

# 或手动修复：删除配置文件
sudo rm /etc/docker/daemon.json
sudo systemctl restart docker
```

### Docker 镜像拉取超时

**错误信息：**

```
Get "https://registry-1.docker.io/v2/": net/http: request canceled
```

**解决方案：**

```bash
# 配置镜像加速器
sudo ./scripts/configure-docker-mirror.sh

# 或预先拉取镜像
./scripts/pre-pull-images.sh
```

### 端口被占用

```bash
# 检查端口占用
sudo ss -tulpn | grep -E '8080|80|3306'

# 停止占用端口的服务或修改docker-compose.yml中的端口映射
```

### 数据库连接失败

```bash
# 检查MySQL容器是否运行
docker ps | grep mysql

# 检查环境变量
cat .env | grep DB_

# 测试数据库连接
docker exec -it blog-mysql mysql -u blog_user -p
```

### 容器无法启动

```bash
# 查看详细错误日志
docker-compose -f docker-compose.prod.yml logs

# 检查容器状态
docker ps -a

# 查看特定容器日志
docker logs blog-api
```

## 联系支持

如遇到问题，请查看：

- [README.md](./README.md) - 项目说明
- [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) - API 接口文档
- [ARCHITECTURE.md](./ARCHITECTURE.md) - 项目架构说明
