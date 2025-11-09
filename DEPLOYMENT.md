# 博客系统部署文档

本文档详细说明如何在Linux系统上使用Docker完成博客系统的自动化部署。

## 系统要求

- Linux操作系统（推荐Ubuntu 20.04+ 或 CentOS 7+）
- Docker 20.10+
- Docker Compose 2.0+
- 至少2GB可用内存
- 至少10GB可用磁盘空间

## 架构说明

系统采用前后端分离架构：

- **后端**: Golang API服务（端口8080）
- **前端**: Vue.js应用，通过Nginx提供服务（端口80/443）
- **数据库**: MySQL 8.0（端口3306）

所有服务通过Docker容器运行，使用镜像加速器加速镜像拉取。

## 部署步骤

### 第一步：安装Docker和Docker Compose

#### Ubuntu/Debian系统

```bash
# 更新系统包
sudo apt-get update

# 安装必要的依赖
sudo apt-get install -y ca-certificates curl gnupg lsb-release

# 添加Docker官方GPG密钥
sudo mkdir -p /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg

# 设置Docker仓库
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
  $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# 安装Docker Engine和Docker Compose
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 验证安装
docker --version
docker compose version
```

#### CentOS/RHEL系统

```bash
# 安装必要的依赖
sudo yum install -y yum-utils

# 添加Docker仓库
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo

# 安装Docker Engine和Docker Compose
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# 启动Docker服务
sudo systemctl start docker
sudo systemctl enable docker

# 验证安装
docker --version
docker compose version
```

### 第二步：配置Docker镜像加速器

为了加快镜像拉取速度，配置镜像加速器：

```bash
# 创建Docker配置目录
sudo mkdir -p /etc/docker

# 配置镜像加速器
sudo tee /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.1ms.run/library"
  ],
  "max-concurrent-downloads": 10,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

# 重启Docker服务
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 第三步：准备项目文件

#### 3.1 克隆或上传项目代码

```bash
# 方式1: 使用Git克隆
git clone <your-repo-url> /opt/blog
cd /opt/blog/api

# 方式2: 上传项目文件到服务器
# 使用scp或FTP工具上传项目文件到服务器
```

#### 3.2 配置环境变量

```bash
# 从模板创建.env文件
cp env.template .env

# 编辑.env文件，配置数据库等信息
vi .env
```

`.env`文件配置示例：

```env
# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=your_secure_password
DB_NAME=blog

# MySQL Root密码
MYSQL_ROOT_PASSWORD=your_root_password

# API基础URL（用于生成图片完整URL）
API_BASE_URL=http://your-domain.com

# 环境标识
BLOG_ENV=prod
```

**重要提示**：
- 生产环境请使用强密码
- `API_BASE_URL`应设置为实际的域名或IP地址
- 确保密码安全，不要将`.env`文件提交到Git仓库

### 第四步：准备前端文件

确保Vue前端项目已经构建完成：

```bash
# 在前端项目目录中构建
cd /path/to/front
npm install
npm run build

# 构建完成后，dist目录应该在 ../front/dist
```

如果前端文件不在`../front/dist`目录，需要修改`docker-compose.yml`中的前端挂载路径。

### 第五步：执行部署

#### 方式1：直接部署（推荐，自动拉取镜像）

```bash
# 进入项目目录
cd /opt/blog/api

# 运行部署脚本
sudo ./scripts/deploy.sh production
```

部署脚本会自动：
1. 检查Docker和Docker Compose
2. 配置Docker镜像加速器
3. 检查环境配置文件
4. 停止旧容器
5. 构建后端镜像
6. 启动所有服务

#### 方式2：离线部署（使用预打包的镜像）

如果服务器网络受限，可以预先打包镜像：

**在本地或网络良好的机器上：**

```bash
# 运行打包脚本
./scripts/package.sh

# 脚本会生成 docker-images.tar.gz 文件
```

**在服务器上：**

```bash
# 上传镜像包到服务器
scp docker-images.tar.gz user@server:/opt/blog/api/

# 在服务器上加载镜像
cd /opt/blog/api
tar -xzf docker-images.tar.gz
cd docker-package/images
for image in *.tar; do
    docker load -i "$image"
done
cd ../../

# 运行部署脚本
sudo ./scripts/deploy.sh production
```

### 第六步：验证部署

#### 6.1 检查服务状态

```bash
# 查看所有容器状态
docker compose -f docker-compose.prod.yml ps

# 应该看到三个容器都在运行：
# - blog-mysql
# - blog-api
# - blog-frontend
```

#### 6.2 查看服务日志

```bash
# 查看所有服务日志
docker compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker compose -f docker-compose.prod.yml logs -f api
docker compose -f docker-compose.prod.yml logs -f mysql
docker compose -f docker-compose.prod.yml logs -f frontend
```

#### 6.3 测试服务

```bash
# 测试API服务
curl http://localhost:8080/health

# 测试前端服务
curl http://localhost

# 测试数据库连接
docker exec -it blog-mysql mysql -u blog_user -p -e "SHOW DATABASES;"
```

### 第七步：配置Nginx（可选）

如果需要配置HTTPS或自定义Nginx配置：

```bash
# 创建Nginx配置目录（如果不存在）
mkdir -p docker/nginx/conf.d

# 创建自定义Nginx配置
vi docker/nginx/conf.d/default.conf
```

Nginx配置示例：

```nginx
server {
    listen 80;
    server_name your-domain.com;

    # 前端静态文件
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    # API代理
    location /api {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # 上传文件访问
    location /uploads {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
    }
}
```

配置完成后，重启前端服务：

```bash
docker compose -f docker-compose.prod.yml restart frontend
```

## 常用操作命令

### 查看服务状态

```bash
docker compose -f docker-compose.prod.yml ps
```

### 查看日志

```bash
# 查看所有服务日志
docker compose -f docker-compose.prod.yml logs -f

# 查看最近100行日志
docker compose -f docker-compose.prod.yml logs --tail=100

# 查看特定服务日志
docker compose -f docker-compose.prod.yml logs -f api
```

### 重启服务

```bash
# 重启所有服务
docker compose -f docker-compose.prod.yml restart

# 重启特定服务
docker compose -f docker-compose.prod.yml restart api
```

### 停止服务

```bash
docker compose -f docker-compose.prod.yml down
```

### 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建并启动
docker compose -f docker-compose.prod.yml up -d --build
```

### 进入容器

```bash
# 进入API容器
docker exec -it blog-api /bin/bash

# 进入MySQL容器
docker exec -it blog-mysql mysql -u root -p

# 进入前端容器
docker exec -it blog-frontend /bin/bash
```

### 备份数据库

```bash
# 备份数据库
docker exec blog-mysql mysqldump -u root -p${MYSQL_ROOT_PASSWORD} blog > backup_$(date +%Y%m%d_%H%M%S).sql

# 恢复数据库
docker exec -i blog-mysql mysql -u root -p${MYSQL_ROOT_PASSWORD} blog < backup.sql
```

## 故障排查

### 问题1：容器无法启动

**检查方法：**

```bash
# 查看容器日志
docker compose -f docker-compose.prod.yml logs

# 查看容器状态
docker compose -f docker-compose.prod.yml ps -a
```

**常见原因：**
- 环境变量配置错误
- 端口被占用
- 磁盘空间不足
- 内存不足

### 问题2：数据库连接失败

**检查方法：**

```bash
# 检查MySQL容器是否运行
docker ps | grep mysql

# 检查MySQL日志
docker logs blog-mysql

# 测试数据库连接
docker exec -it blog-mysql mysql -u blog_user -p -e "SELECT 1;"
```

**解决方案：**
- 确认`.env`文件中的数据库配置正确
- 等待MySQL完全启动（首次启动需要初始化）
- 检查网络连接

### 问题3：API服务无法访问

**检查方法：**

```bash
# 检查API容器日志
docker logs blog-api

# 检查API容器是否运行
docker ps | grep api

# 测试API端点
curl http://localhost:8080/health
```

**解决方案：**
- 检查端口是否被占用：`netstat -tlnp | grep 8080`
- 查看API日志找出错误原因
- 确认数据库连接正常

### 问题4：前端页面无法访问

**检查方法：**

```bash
# 检查前端容器日志
docker logs blog-frontend

# 检查前端文件是否存在
docker exec blog-frontend ls -la /usr/share/nginx/html
```

**解决方案：**
- 确认前端构建文件在正确位置
- 检查Nginx配置是否正确
- 确认端口80/443未被占用

### 问题5：镜像拉取失败

**解决方案：**

```bash
# 检查Docker镜像加速器配置
cat /etc/docker/daemon.json

# 重启Docker服务
sudo systemctl restart docker

# 手动拉取镜像
docker pull golang:latest
```

## 性能优化建议

### 1. 数据库优化

```bash
# 在docker-compose.prod.yml中添加MySQL配置
command: >
  --default-authentication-plugin=mysql_native_password
  --character-set-server=utf8mb4
  --collation-server=utf8mb4_unicode_ci
  --innodb-buffer-pool-size=1G
  --max-connections=500
```

### 2. 资源限制

在`docker-compose.prod.yml`中为服务添加资源限制：

```yaml
services:
  api:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

### 3. 日志管理

定期清理Docker日志：

```bash
# 清理所有停止的容器
docker container prune -f

# 清理未使用的镜像
docker image prune -a -f

# 清理未使用的卷
docker volume prune -f
```

## 安全建议

1. **使用强密码**：生产环境必须使用强密码
2. **限制端口访问**：使用防火墙限制数据库端口（3306）的外部访问
3. **定期更新**：定期更新Docker镜像和系统补丁
4. **备份数据**：定期备份数据库和上传的文件
5. **HTTPS配置**：生产环境建议配置HTTPS证书

## 维护计划

### 日常维护

- 每日检查服务状态
- 每周查看日志文件
- 每月备份数据库

### 定期维护

- 每季度更新Docker镜像
- 每半年审查安全配置
- 每年更新系统依赖

## 技术支持

如遇到问题，请：

1. 查看本文档的故障排查部分
2. 检查服务日志
3. 查看项目Issue或联系技术支持

---

**最后更新**: 2024年
