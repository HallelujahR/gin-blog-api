# CentOS 服务器部署执行顺序

## 📋 部署流程图

```
服务器准备 → 安装Docker → 克隆代码 → 配置环境 → 首次部署 → 验证测试
```

## 🚀 详细执行步骤

### 步骤 1: 服务器准备（5 分钟）

#### 1.1 登录服务器

```bash
ssh root@your-server-ip
# 或使用普通用户
ssh user@your-server-ip
```

#### 1.2 更新系统

```bash
# CentOS 7
sudo yum update -y

# CentOS 8/9 或 Rocky Linux
sudo dnf update -y
```

#### 1.3 安装必要工具

```bash
# CentOS 7
sudo yum install -y git curl wget vim

# CentOS 8/9
sudo dnf install -y git curl wget vim
```

#### 1.4 配置防火墙（开放必要端口）

```bash
# 检查防火墙状态
sudo systemctl status firewalld

# 如果防火墙开启，开放端口
sudo firewall-cmd --permanent --add-port=80/tcp
sudo firewall-cmd --permanent --add-port=443/tcp
sudo firewall-cmd --permanent --add-port=8080/tcp
sudo firewall-cmd --permanent --add-port=22/tcp
sudo firewall-cmd --reload

# 查看开放的端口
sudo firewall-cmd --list-ports
```

---

### 步骤 2: 安装 Docker（10 分钟）

#### 2.1 卸载旧版本（如果有）

```bash
sudo yum remove docker docker-client docker-client-latest \
    docker-common docker-latest docker-latest-logrotate \
    docker-logrotate docker-engine
```

#### 2.2 安装依赖包

```bash
# CentOS 7
sudo yum install -y yum-utils device-mapper-persistent-data lvm2

# CentOS 8/9
sudo dnf install -y yum-utils device-mapper-persistent-data lvm2
```

#### 2.3 添加 Docker 官方仓库

```bash
# CentOS 7
sudo yum-config-manager --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo

# CentOS 8/9
sudo dnf config-manager --add-repo \
    https://download.docker.com/linux/centos/docker-ce.repo
```

#### 2.4 安装 Docker Engine

```bash
# CentOS 7
sudo yum install docker

# CentOS 8/9
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

#### 2.5 启动 Docker 服务

```bash
# 启动Docker
sudo systemctl start docker

# 设置开机自启
sudo systemctl enable docker

# 验证安装
sudo docker --version
```

#### 2.6 安装 Docker Compose（如果未包含在 docker-compose-plugin 中）

```bash
# 方式1: 使用官方脚本（推荐）
sudo yum install docker-compose

# 验证安装
docker-compose --version
# 或
docker compose version
```

#### 2.7 配置 Docker 用户组（可选，避免每次使用 sudo）

```bash
# 将当前用户添加到docker组
sudo usermod -aG docker $USER

# 重新登录或执行以下命令使组更改生效
newgrp docker

# 测试（不需要sudo）
docker ps
```

---

### 步骤 3: 克隆代码（5 分钟）

#### 3.1 创建项目目录

```bash
# 创建目录
sudo mkdir -p /opt/blog
sudo chown $USER:$USER /opt/blog
cd /opt/blog
```

#### 3.2 克隆后端代码

```bash
# 替换为你的GitHub仓库地址
git clone https://github.com/HallelujahR/gin-blog-api.git api
cd api
```

#### 3.3 克隆前端代码（如果前端在单独仓库）

```bash
cd /opt/blog
# 替换为你的前端仓库地址
git clone https://github.com/your-username/blog-front.git front
```

**如果前端和后端在同一仓库的不同目录：**

```bash
# 如果前端代码在 ../front 目录
# 确保目录结构正确
ls -la /opt/blog/
```

---

### 步骤 4: 配置环境变量（10 分钟）

#### 4.1 创建环境变量文件

```bash
cd /opt/blog/api

# 复制模板
cp env.template .env

# 编辑环境变量
vi .env
# 或使用 nano
nano .env
```

#### 4.2 配置 .env 文件内容

```env
# 数据库配置
DB_HOST=mysql
DB_PORT=3306
DB_USER=blog_user
DB_PASSWORD=your_strong_password_here
DB_NAME=blog

# MySQL Root密码
MYSQL_ROOT_PASSWORD=your_root_password_here

# API基础URL（替换为你的实际域名或IP）
# 如果有域名: API_BASE_URL=https://api.yourdomain.com
# 如果只有IP: API_BASE_URL=http://your-server-ip:8080
API_BASE_URL=http://your-server-ip:8080

# 环境标识
BLOG_ENV=prod
```

**重要提示：**

- `DB_PASSWORD` 和 `MYSQL_ROOT_PASSWORD` 必须设置强密码
- `API_BASE_URL` 如果使用域名，确保 DNS 已解析
- 如果只有 IP，使用 `http://your-server-ip:8080`

#### 4.3 配置前端 API 地址

```bash
# 编辑前端API配置文件
cd /opt/blog/front
vi src/api/index.js
# 或
nano src/api/index.js
```

**修改内容：**

```javascript
const http = axios.create({
  // 如果有域名
  baseURL: "https://api.yourdomain.com/api",
  // 如果只有IP
  // baseURL: 'http://your-server-ip:8080/api',
  timeout: 7000,
});
```

---

### 步骤 5: 首次部署（15 分钟）

#### 5.1 进入后端目录

```bash
cd /opt/blog/api
```

#### 5.2 给脚本添加执行权限

```bash
chmod +x scripts/*.sh
```

#### 5.3 运行部署脚本

```bash
# 使用生产环境配置部署
./scripts/deploy.sh production
```

**脚本执行过程：**

1. 检查 Docker 和 Docker Compose
2. 检查.env 文件
3. 拉取最新代码（如果有 Git）
4. 停止旧容器
5. 构建 Docker 镜像
6. 启动服务（API、MySQL、前端）

**预计时间：**

- 首次构建：10-15 分钟（下载镜像和编译）
- 后续部署：3-5 分钟

#### 5.4 查看部署日志

```bash
# 查看所有服务日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker-compose -f docker-compose.prod.yml logs -f api
docker-compose -f docker-compose.prod.yml logs -f mysql
docker-compose -f docker-compose.prod.yml logs -f frontend
```

---

### 步骤 6: 验证部署（5 分钟）

#### 6.1 检查容器状态

```bash
# 查看所有容器状态
docker-compose -f docker-compose.prod.yml ps

# 应该看到三个容器都在运行：
# - blog-api (后端)
# - blog-mysql (数据库)
# - blog-frontend (前端)
```

#### 6.2 测试 API 接口

```bash
# 测试文章列表接口
curl http://localhost:8080/api/posts?page=1&size=1

# 测试分类接口
curl http://localhost:8080/api/categories

# 如果返回JSON数据，说明API正常
```

#### 6.3 测试前端页面

```bash
# 测试前端页面
curl http://localhost

# 应该返回HTML内容
```

#### 6.4 检查数据库连接

```bash
# 进入MySQL容器
docker exec -it blog-mysql mysql -u blog_user -p

# 输入密码后，执行SQL
SHOW DATABASES;
USE blog;
SHOW TABLES;
EXIT;
```

#### 6.5 浏览器访问测试

```bash
# 如果服务器有公网IP或域名
# 在浏览器访问：
# http://your-server-ip
# 或
# http://your-domain.com
```

---

## 🔄 后续更新流程

### 方式 1: 使用更新脚本（推荐，零停机）

```bash
cd /opt/blog/api
./scripts/update.sh production
```

### 方式 2: 手动更新

```bash
cd /opt/blog/api

# 拉取最新代码
git pull origin main

# 重新构建并启动
docker-compose -f docker-compose.prod.yml up -d --build
```

---

## 🛠️ 常用管理命令

### 查看服务状态

```bash
docker-compose -f docker-compose.prod.yml ps
```

### 查看日志

```bash
# 查看所有日志
docker-compose -f docker-compose.prod.yml logs -f

# 查看最近100行
docker-compose -f docker-compose.prod.yml logs --tail=100
```

### 重启服务

```bash
# 重启所有服务
docker-compose -f docker-compose.prod.yml restart

# 重启特定服务
docker-compose -f docker-compose.prod.yml restart api
```

### 停止服务

```bash
docker-compose -f docker-compose.prod.yml down
```

### 停止并删除数据卷（谨慎使用）

```bash
docker-compose -f docker-compose.prod.yml down -v
```

### 进入容器

```bash
# 进入API容器
docker exec -it blog-api sh

# 进入MySQL容器
docker exec -it blog-mysql bash

# 进入前端容器
docker exec -it blog-frontend sh
```

---

## ⚠️ 常见问题排查

### 问题 1: Docker 安装失败

```bash
# 检查系统版本
cat /etc/centos-release

# 如果是CentOS 7，确保已安装EPEL仓库
sudo yum install -y epel-release

# 清理yum缓存
sudo yum clean all
sudo yum makecache
```

### 问题 2: 端口被占用

```bash
# 检查端口占用
sudo netstat -tulpn | grep -E '8080|80|3306'

# 或使用ss命令
sudo ss -tulpn | grep -E '8080|80|3306'

# 如果端口被占用，停止占用该端口的服务或修改docker-compose.yml中的端口映射
```

### 问题 3: 容器无法启动

```bash
# 查看详细错误日志
docker-compose -f docker-compose.prod.yml logs

# 检查容器状态
docker ps -a

# 查看特定容器的详细信息
docker inspect blog-api
```

### 问题 4: 数据库连接失败

```bash
# 检查MySQL容器是否运行
docker ps | grep mysql

# 查看MySQL日志
docker logs blog-mysql

# 检查环境变量
cat .env | grep DB_

# 测试数据库连接
docker exec -it blog-mysql mysql -u blog_user -p
```

### 问题 5: 磁盘空间不足

```bash
# 查看磁盘使用
df -h

# 清理未使用的Docker资源
docker system prune -a

# 清理未使用的镜像
docker image prune -a
```

---

## 📊 部署时间预估

| 步骤        | 预计时间       | 说明                    |
| ----------- | -------------- | ----------------------- |
| 服务器准备  | 5 分钟         | 更新系统、配置防火墙    |
| 安装 Docker | 10 分钟        | 下载和安装 Docker       |
| 克隆代码    | 5 分钟         | 从 GitHub 克隆代码      |
| 配置环境    | 10 分钟        | 配置环境变量和 API 地址 |
| 首次部署    | 15 分钟        | 构建镜像和启动服务      |
| 验证测试    | 5 分钟         | 测试各项功能            |
| **总计**    | **约 50 分钟** | 首次部署完整流程        |

---

## ✅ 部署检查清单

部署完成后，请确认：

- [ ] Docker 和 Docker Compose 已正确安装
- [ ] 所有容器都在运行（3 个容器）
- [ ] API 接口可以正常访问
- [ ] 前端页面可以正常访问
- [ ] 数据库连接正常
- [ ] 文件上传功能正常
- [ ] 日志无错误信息

---

## 🔗 相关文档

- [DEPLOYMENT.md](./DEPLOYMENT.md) - 完整部署文档
- [DEPLOYMENT_CHECKLIST.md](./DEPLOYMENT_CHECKLIST.md) - 部署检查清单
- [README.md](./README.md) - 项目说明

---

## 💡 提示

1. **首次部署建议**：在非高峰期进行，确保有足够时间处理问题
2. **备份重要**：部署前备份现有数据（如果有）
3. **监控资源**：部署后监控服务器 CPU、内存、磁盘使用情况
4. **日志管理**：定期查看日志，及时发现问题
5. **安全配置**：部署完成后修改所有默认密码，配置 HTTPS
