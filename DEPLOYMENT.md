# 部署文档

## 部署方式

本项目使用从本地打包的Docker安装包和镜像进行部署，避免在服务器上拉取镜像时的网络问题。

## 部署流程

### 1. 本地打包（在本地开发机器上）

#### 1.1 打包Docker和镜像

```bash
# 运行打包脚本
./scripts/package.sh
```

脚本会：
- 检查Docker安装包（需要在 `packages/docker-ce/` 目录）
- 拉取或导出所需的Docker镜像
- 创建安装脚本
- 打包成tar.gz文件

#### 1.2 准备Docker安装包（可选）

如果没有Docker安装包，可以在CentOS机器上下载：

```bash
# 创建目录
mkdir -p packages/docker-ce
cd packages/docker-ce

# 下载Docker安装包（需要先配置yum源）
yum install --downloadonly --downloaddir=. docker-ce docker-ce-cli containerd.io docker-compose-plugin
```

### 2. 上传到服务器

将打包好的文件上传到服务器：

```bash
# 使用scp上传
scp docker-package-*.tar.gz user@server:/path/to/project/
```

### 3. 服务器部署

在服务器上运行部署脚本：

```bash
# 进入项目目录
cd /path/to/project

# 运行部署脚本（需要root权限）
sudo ./scripts/deploy.sh production docker-package-*.tar.gz
```

部署脚本会：
- 从tar包安装Docker（如果未安装）
- 加载Docker镜像
- 配置Docker镜像加速器
- 构建和启动应用

## 环境配置

### 创建.env文件

部署前需要创建 `.env` 文件：

```bash
# 从模板创建
cp env.template .env

# 编辑配置
vi .env
```

### 环境变量说明

- `DB_HOST`: 数据库主机地址
- `DB_PORT`: 数据库端口（默认3306）
- `DB_USER`: 数据库用户名
- `DB_PASSWORD`: 数据库密码
- `DB_NAME`: 数据库名称
- `MYSQL_ROOT_PASSWORD`: MySQL root密码
- `API_BASE_URL`: API服务基础URL

## 常用命令

### 查看服务状态

```bash
docker compose -f docker-compose.prod.yml ps
```

### 查看日志

```bash
# 查看所有服务日志
docker compose -f docker-compose.prod.yml logs -f

# 查看特定服务日志
docker compose -f docker-compose.prod.yml logs -f api
```

### 重启服务

```bash
docker compose -f docker-compose.prod.yml restart
```

### 停止服务

```bash
docker compose -f docker-compose.prod.yml down
```

### 更新部署

```bash
# 拉取最新代码
git pull origin main

# 重新构建和启动
docker compose -f docker-compose.prod.yml up -d --build
```

## 故障排查

### Docker服务未启动

```bash
# 检查Docker服务状态
systemctl status docker

# 启动Docker服务
systemctl start docker
```

### 镜像加载失败

```bash
# 手动加载镜像
docker load -i docker-package/images/golang_1.22.tar
```

### 查看Docker日志

```bash
# 查看Docker服务日志
journalctl -u docker.service -n 50
```

## 注意事项

1. 部署脚本需要root权限
2. 确保服务器有足够的磁盘空间
3. 确保网络连接正常（用于拉取代码）
4. 首次部署需要创建.env配置文件
