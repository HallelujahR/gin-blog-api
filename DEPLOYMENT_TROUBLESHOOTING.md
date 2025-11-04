# 部署问题排查指南

## Docker镜像拉取超时

### 错误信息
```
ERROR: Service 'api' failed to build: Get https://registry-1.docker.io/v2/: 
net/http: request canceled while waiting for connection 
(Client.Timeout exceeded while awaiting headers)
```

### 原因
- Docker Hub连接不稳定或无法访问
- 网络防火墙限制
- 在中国大陆访问Docker Hub速度慢

### 解决方案

#### 方案1：配置Docker镜像加速器（推荐）

```bash
# 使用项目提供的配置脚本
cd /opt/blog/api
chmod +x scripts/configure-docker-mirror.sh
sudo ./scripts/configure-docker-mirror.sh
```

**手动配置：**
```bash
# 创建配置目录
sudo mkdir -p /etc/docker

# 创建/编辑daemon.json
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com",
    "https://registry.docker-cn.com"
  ],
  "max-concurrent-downloads": 10
}
EOF

# 重启Docker
sudo systemctl daemon-reload
sudo systemctl restart docker

# 验证配置
docker info | grep -A 10 "Registry Mirrors"
```

#### 方案2：手动拉取镜像

```bash
# 先拉取需要的镜像
docker pull golang:1.23-alpine
docker pull mysql:8.0
docker pull nginx:alpine
docker pull node:18-alpine

# 然后再执行部署
./scripts/deploy.sh production
```

#### 方案3：使用代理

如果有HTTP代理，可以配置：

```bash
# 创建Docker客户端配置目录
mkdir -p ~/.docker

# 配置代理
cat > ~/.docker/config.json <<EOF
{
  "proxies": {
    "default": {
      "httpProxy": "http://proxy.example.com:8080",
      "httpsProxy": "http://proxy.example.com:8080",
      "noProxy": "localhost,127.0.0.1"
    }
  }
}
EOF
```

#### 方案4：增加超时时间

```bash
# 编辑Docker配置
sudo tee -a /etc/docker/daemon.json > /dev/null <<EOF
{
  "max-concurrent-downloads": 5,
  "max-download-attempts": 3
}
EOF

sudo systemctl restart docker
```

---

## Docker Compose版本不兼容

### 错误信息
```
Version in "./docker-compose.prod.yml" is unsupported
```

### 解决方案
已修复，使用 `version: '3.3'` 兼容旧版Docker Compose。

---

## Go版本不存在

### 错误信息
```
Error parsing reference: "golang:1.24-alpine AS builder" is not a valid repository/tag
```

### 解决方案
已修复，使用 `golang:1.23-alpine` 稳定版本。

---

## 多阶段构建不支持

### 错误信息
```
"golang:1.23-alpine AS builder" is not a valid repository/tag: invalid reference format
```

### 解决方案
已修复，使用单阶段构建兼容旧版Docker。

---

## 端口被占用

### 错误信息
```
Error starting userland proxy: listen tcp4 0.0.0.0:8080: bind: address already in use
```

### 解决方案
```bash
# 检查端口占用
sudo netstat -tulpn | grep -E '8080|80|3306'

# 或使用ss命令
sudo ss -tulpn | grep -E '8080|80|3306'

# 停止占用端口的服务
sudo kill -9 <PID>

# 或修改docker-compose.yml中的端口映射
```

---

## 数据库连接失败

### 错误信息
```
数据库连接失败: dial tcp: lookup mysql on 127.0.0.11:53: no such host
```

### 解决方案
```bash
# 检查MySQL容器是否运行
docker ps | grep mysql

# 检查环境变量
cat .env | grep DB_

# 确保.env文件中的DB_HOST=mysql（Docker Compose网络）
# 确保MySQL容器先启动
docker-compose -f docker-compose.prod.yml up -d mysql
sleep 10
docker-compose -f docker-compose.prod.yml up -d api
```

---

## 权限问题

### 错误信息
```
permission denied while trying to connect to the Docker daemon socket
```

### 解决方案
```bash
# 将用户添加到docker组
sudo usermod -aG docker $USER

# 重新登录或执行
newgrp docker

# 验证
docker ps
```

---

## 磁盘空间不足

### 错误信息
```
no space left on device
```

### 解决方案
```bash
# 查看磁盘使用
df -h

# 清理Docker资源
docker system prune -a -f

# 清理未使用的镜像
docker image prune -a -f

# 清理未使用的卷
docker volume prune -f

# 查看Docker占用空间
docker system df
```

---

## 容器无法启动

### 排查步骤
```bash
# 1. 查看容器状态
docker-compose -f docker-compose.prod.yml ps -a

# 2. 查看容器日志
docker-compose -f docker-compose.prod.yml logs api
docker-compose -f docker-compose.prod.yml logs mysql

# 3. 查看容器详细信息
docker inspect blog-api

# 4. 进入容器检查
docker exec -it blog-api sh
```

---

## 网络连接问题

### 检查网络
```bash
# 检查Docker网络
docker network ls
docker network inspect blog-network

# 检查容器网络连接
docker exec -it blog-api ping mysql
docker exec -it blog-api ping google.com
```

---

## 常见问题快速诊断

```bash
# 1. 检查Docker版本
docker --version
docker compose version

# 2. 检查Docker服务状态
sudo systemctl status docker

# 3. 检查镜像加速器
docker info | grep -A 10 "Registry Mirrors"

# 4. 检查容器状态
docker ps -a

# 5. 检查日志
docker-compose -f docker-compose.prod.yml logs --tail=50

# 6. 检查环境变量
cat .env

# 7. 检查端口占用
sudo ss -tulpn | grep -E '8080|80|3306'

# 8. 检查磁盘空间
df -h
docker system df
```

---

## 获取帮助

如果以上方案都无法解决问题：

1. 查看详细日志：`docker-compose logs -f`
2. 检查系统资源：`top`, `free -h`, `df -h`
3. 查看Docker信息：`docker info`
4. 查看系统日志：`journalctl -u docker`

