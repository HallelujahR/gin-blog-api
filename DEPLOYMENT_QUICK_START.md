# CentOS 快速部署指南（一键执行）

## 🚀 快速部署脚本（推荐）

### 一键执行所有步骤

```bash
# 1. 创建并进入工作目录
sudo mkdir -p /opt/blog
sudo chown $USER:$USER /opt/blog
cd /opt/blog

# 2. 下载并执行部署脚本
curl -fsSL https://raw.githubusercontent.com/your-repo/gin-blog-api/main/scripts/centos-deploy.sh -o centos-deploy.sh
chmod +x centos-deploy.sh
./centos-deploy.sh
```

---

## 📝 手动执行步骤（详细版）

### 第一步：安装Docker（必须）

```bash
# CentOS 7
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl start docker
sudo systemctl enable docker

# CentOS 8/9
sudo dnf install -y yum-utils
sudo dnf config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo dnf install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
sudo systemctl start docker
sudo systemctl enable docker

# 验证
docker --version
docker compose version
```

### 第二步：克隆代码（必须）

```bash
cd /opt/blog
git clone https://github.com/HallelujahR/gin-blog-api.git api
cd api
```

### 第三步：配置环境（必须）

```bash
# 复制环境变量模板
cp env.template .env

# 编辑环境变量（必须修改密码和API地址）
vi .env
```

**最少需要修改的内容：**
```env
DB_PASSWORD=你的数据库密码
MYSQL_ROOT_PASSWORD=你的MySQL root密码
API_BASE_URL=http://你的服务器IP:8080
```

### 第四步：部署（必须）

```bash
# 添加执行权限
chmod +x scripts/*.sh

# 执行部署
./scripts/deploy.sh production
```

### 第五步：验证（必须）

```bash
# 检查容器状态
docker-compose -f docker-compose.prod.yml ps

# 测试API
curl http://localhost:8080/api/posts?page=1&size=1

# 测试前端
curl http://localhost
```

---

## ⏱️ 执行时间线

```
00:00 - 00:10  安装Docker
00:10 - 00:15  克隆代码
00:15 - 00:25  配置环境变量
00:25 - 00:40  首次部署（构建镜像）
00:40 - 00:45  验证测试
```

**总时间：约45分钟**

---

## 🔍 快速检查命令

```bash
# 检查Docker
docker --version && docker compose version

# 检查容器
docker-compose -f docker-compose.prod.yml ps

# 检查日志
docker-compose -f docker-compose.prod.yml logs --tail=50

# 检查端口
sudo ss -tulpn | grep -E '8080|80|3306'
```

---

## 🆘 遇到问题？

1. **查看详细文档**：[DEPLOYMENT_CENTOS.md](./DEPLOYMENT_CENTOS.md)
2. **查看日志**：`docker-compose logs -f`
3. **检查状态**：`docker-compose ps`

---

## 📌 重要提示

1. ⚠️ **必须修改** `.env` 文件中的密码
2. ⚠️ **必须配置** 防火墙开放端口
3. ⚠️ **确保** 服务器有足够内存（至少2GB）
4. ⚠️ **首次部署** 需要下载镜像，可能需要10-15分钟

