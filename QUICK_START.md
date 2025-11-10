# 快速部署指南

本文档说明如何使用服务器上的本地镜像快速部署 Golang 博客系统。

## 前提条件

服务器上需要已安装以下 Docker 镜像：

- `docker.1ms.run/library/golang:latest`
- `docker.1ms.run/library/mysql:8.0.44`
- `docker.1ms.run/library/nginx:latest`
- `docker.1ms.run/library/node:latest` (可选，用于前端构建)

## 快速部署步骤

### 1. 检查镜像是否存在

```bash
docker images | grep "docker.1ms.run/library"
```

应该看到以下镜像：
- docker.1ms.run/library/golang:latest
- docker.1ms.run/library/mysql:8.0.44
- docker.1ms.run/library/nginx:latest
- docker.1ms.run/library/node:latest

### 2. 准备环境配置

```bash
# 从模板创建.env文件
cp env.template .env

# 编辑配置文件
vi .env
```

配置示例：
```env
DB_HOST=mysql
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=your_password
DB_NAME=blog
MYSQL_ROOT_PASSWORD=root_password
API_BASE_URL=http://your-domain.com
BLOG_ENV=prod
```

### 3. 执行部署

```bash
# 生产环境部署
sudo ./scripts/deploy.sh production

# 或开发环境部署
sudo ./scripts/deploy.sh development
```

部署脚本会自动：
1. 检查 Docker 和 Docker Compose
2. 检查必需的镜像是否存在
3. 检查环境配置文件
4. 停止旧容器
5. 构建应用镜像（使用本地基础镜像）
6. 启动所有服务

### 4. 验证部署

```bash
# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f

# 测试API
curl http://localhost:8080/health
```

## 镜像打包（可选）

如果需要将镜像打包到其他服务器：

```bash
# 打包所有镜像
./scripts/package.sh

# 会生成 docker-images.tar.gz 文件
```

在其他服务器上加载镜像：

```bash
# 解压镜像包
tar -xzf docker-images.tar.gz

# 加载镜像
cd docker-package/images
for image in *.tar; do
    docker load -i "$image"
done
```

## 常用命令

```bash
# 查看服务状态
docker compose -f docker-compose.prod.yml ps

# 查看日志
docker compose -f docker-compose.prod.yml logs -f api
docker compose -f docker-compose.prod.yml logs -f mysql

# 重启服务
docker compose -f docker-compose.prod.yml restart

# 停止服务
docker compose -f docker-compose.prod.yml down

# 更新代码后重新部署
git pull
sudo ./scripts/deploy.sh production
```

## 故障排查

### 镜像不存在

如果提示镜像不存在，请先加载镜像：

```bash
# 如果有镜像包
docker load -i <镜像包路径>

# 或从镜像源拉取（如果网络允许）
docker pull docker.1ms.run/library/golang:latest
docker pull docker.1ms.run/library/mysql:8.0.44
docker pull docker.1ms.run/library/nginx:latest
```

### 端口被占用

```bash
# 检查端口占用
netstat -tlnp | grep 8080
netstat -tlnp | grep 3306

# 停止占用端口的服务或修改 docker-compose.yml 中的端口映射
```

### 数据库连接失败

```bash
# 检查MySQL容器日志
docker logs blog-mysql

# 检查.env文件中的数据库配置是否正确
cat .env | grep DB_
```

## 注意事项

1. 所有脚本使用 `pull_policy: never`，确保只使用本地镜像
2. 构建时使用 `--pull=false`，不会尝试拉取新镜像
3. 确保服务器有足够的磁盘空间和内存
4. 生产环境请使用强密码

