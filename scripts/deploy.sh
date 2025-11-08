#!/bin/bash

# 博客系统自动部署脚本（从tar包部署）
# 功能：从本地tar包安装Docker和镜像，然后部署应用
# 使用方法: sudo ./scripts/deploy.sh [production|staging] [docker-package-path]

set -e

ENV=${1:-production}
DOCKER_PACKAGE=${2:-docker-package.tar.gz}
COMPOSE_FILE="docker-compose.yml"

if [ "$ENV" = "production" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
fi

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用sudo运行此脚本"
    exit 1
fi

echo "🚀 开始部署博客系统 (环境: $ENV)..."
echo "📋 服务器信息:"
echo "   - 操作系统: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'"' -f2 || uname -s)"
echo "   - 内核版本: $(uname -r)"

# 检测操作系统
if [ -f /etc/redhat-release ]; then
    OS_TYPE="centos"
    echo "✅ 检测到CentOS系统"
elif [ -f /etc/debian_version ]; then
    OS_TYPE="debian"
    echo "✅ 检测到Deiban/Ubuntu系统"
else
    echo "⚠️  未知操作系统，假设为CentOS"
    OS_TYPE="centos"
fi

# ========== 安装Docker（从tar包）==========
echo ""
echo "🔧 检查Docker安装状态..."
if ! command -v docker &> /dev/null; then
    echo "📦 从tar包安装Docker..."
    
    # 检查Docker安装包是否存在
    if [ ! -f "$DOCKER_PACKAGE" ]; then
        echo "❌ Docker安装包不存在: $DOCKER_PACKAGE"
        echo "💡 请先运行本地打包脚本: ./scripts/package.sh"
        echo "💡 然后将打包文件上传到服务器"
        exit 1
    fi
    
    # 解压Docker安装包
    echo "📦 解压Docker安装包..."
    EXTRACT_DIR=$(mktemp -d)
    tar -xzf "$DOCKER_PACKAGE" -C "$EXTRACT_DIR"
    DOCKER_PACKAGE_DIR="$EXTRACT_DIR/docker-package"
    
    if [ ! -d "$DOCKER_PACKAGE_DIR" ]; then
        echo "❌ Docker安装包格式错误"
        exit 1
    fi
    
    # 运行安装脚本
    if [ -f "$DOCKER_PACKAGE_DIR/install.sh" ]; then
        echo "🚀 运行Docker安装脚本..."
        chmod +x "$DOCKER_PACKAGE_DIR/install.sh"
        "$DOCKER_PACKAGE_DIR/install.sh"
    else
        echo "❌ 未找到安装脚本"
        exit 1
    fi
    
    # 清理临时目录
    rm -rf "$EXTRACT_DIR"
    
    echo "✅ Docker安装完成"
else
    DOCKER_VERSION=$(docker --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo "✅ Docker已安装，版本: $DOCKER_VERSION"
fi

# ========== 配置Docker镜像加速器 ==========
echo ""
echo "🔧 配置Docker镜像加速器..."
mkdir -p /etc/docker

# 检查是否已有配置且包含镜像加速器
NEED_RESTART=false
if [ -f /etc/docker/daemon.json ]; then
    # 备份现有配置
    cp /etc/docker/daemon.json /etc/docker/daemon.json.bak.$(date +%Y%m%d_%H%M%S)
    # 检查是否已配置镜像加速器
    if ! grep -q "registry-mirrors" /etc/docker/daemon.json; then
        NEED_RESTART=true
    fi
else
    NEED_RESTART=true
fi

# 创建或更新daemon.json
cat > /etc/docker/daemon.json <<'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com",
    "https://registry.docker-cn.com"
  ],
  "max-concurrent-downloads": 10,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

# 重启Docker服务（如果需要）
if [ "$NEED_RESTART" = "true" ]; then
    echo "🔄 重启Docker服务使镜像加速器生效..."
    systemctl daemon-reload
    systemctl restart docker
    
    # 等待Docker服务完全启动
    echo "⏳ 等待Docker服务启动..."
    sleep 5
    
    # 验证Docker是否正常运行
    RETRY=0
    while [ $RETRY -lt 10 ]; do
        if docker info > /dev/null 2>&1; then
            break
        fi
        echo "⏳ 等待Docker服务就绪... ($((RETRY+1))/10)"
        sleep 2
        RETRY=$((RETRY+1))
    done
    
    if ! docker info > /dev/null 2>&1; then
        echo "❌ Docker服务启动失败，请检查日志: journalctl -u docker.service"
        exit 1
    fi
fi

# 验证镜像加速器配置
echo "🔍 验证Docker镜像加速器配置..."
if docker info 2>/dev/null | grep -q "Registry Mirrors"; then
    echo "✅ Docker镜像加速器配置成功"
    docker info 2>/dev/null | grep -A 10 "Registry Mirrors" | head -5
else
    echo "⚠️  无法验证镜像加速器配置，但将继续执行"
fi

# ========== 检测Docker Compose ==========
echo ""
echo "🔧 检查Docker Compose..."
DOCKER_COMPOSE_CMD=""
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
    echo "✅ 使用Docker Compose插件: docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
    echo "✅ 使用Docker Compose独立命令: docker-compose"
else
    echo "⚠️  Docker Compose未安装，尝试安装..."
    if [ "$OS_TYPE" = "centos" ]; then
        # 下载docker-compose
        curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
        DOCKER_COMPOSE_CMD="docker-compose"
    fi
fi

COMPOSE_VERSION=$($DOCKER_COMPOSE_CMD version --short 2>/dev/null || echo "unknown")
echo "📋 Docker Compose版本: $COMPOSE_VERSION"

# ========== 检查.env文件 ==========
echo ""
echo "🔍 检查环境配置文件..."
if [ ! -f .env ]; then
    echo "⚠️  .env文件不存在，正在从模板创建..."
    if [ -f env.template ]; then
        cp env.template .env
        echo "✅ 已创建.env文件，请编辑配置数据库和API地址"
        echo "   编辑命令: vi .env"
        exit 1
    elif [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ 已创建.env文件，请编辑配置数据库和API地址"
        echo "   编辑命令: vi .env"
        exit 1
    else
        echo "❌ 找不到环境变量模板文件"
        exit 1
    fi
fi

# ========== 拉取最新代码 ==========
if [ -d .git ]; then
    echo ""
    echo "📥 拉取最新代码..."
    git pull origin main || echo "⚠️  Git pull失败，继续使用当前代码"
fi

# ========== 检查Docker镜像 ==========
echo ""
echo "🔍 检查必需的Docker镜像..."
REQUIRED_IMAGES=(
    "golang:1.25-alpine"
    "mysql:8.0.44"
    "nginx:latest"
    "node:latest"
)

MISSING_IMAGES=()
for image in "${REQUIRED_IMAGES[@]}"; do
    if docker images "$image" --format "{{.Repository}}:{{.Tag}}" | grep -q "$image"; then
        echo "✅ 镜像已存在: $image"
    else
        echo "⚠️  镜像不存在: $image"
        MISSING_IMAGES+=("$image")
    fi
done

# 如果缺少镜像，尝试从tar包加载
if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    echo ""
    echo "📥 尝试从tar包加载缺失的镜像..."
    if [ -f "$DOCKER_PACKAGE" ]; then
        EXTRACT_DIR=$(mktemp -d)
        tar -xzf "$DOCKER_PACKAGE" -C "$EXTRACT_DIR"
        IMAGES_DIR="$EXTRACT_DIR/docker-package/images"
        
        if [ -d "$IMAGES_DIR" ]; then
            for image_tar in "$IMAGES_DIR"/*.tar; do
                if [ -f "$image_tar" ]; then
                    echo "📥 加载镜像: $(basename $image_tar)"
                    docker load -i "$image_tar" || {
                        echo "⚠️  镜像加载失败: $(basename $image_tar)"
                        continue
                    }
                fi
            done
        fi
        rm -rf "$EXTRACT_DIR"
    else
        echo "⚠️  未找到Docker安装包，无法加载镜像"
        echo "💡 缺失的镜像将在构建时拉取"
    fi
fi

# ========== 停止旧容器 ==========
echo ""
echo "🛑 停止旧容器..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down || true

# ========== 构建镜像 ==========
echo ""
echo "🔨 构建Docker镜像..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE build --no-cache

# ========== 启动服务 ==========
echo ""
echo "🚀 启动服务..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up -d

# ========== 等待服务启动 ==========
echo ""
echo "⏳ 等待服务启动..."
sleep 10

# ========== 检查服务状态 ==========
echo ""
echo "📊 检查服务状态..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps

# ========== 显示日志 ==========
echo ""
echo "📋 最近日志:"
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs --tail=50

echo ""
echo "✅ 部署完成！"
echo ""
echo "📝 服务地址:"
echo "   - 前端: http://your-domain.com"
echo "   - API: http://your-domain.com:8080"
echo ""
echo "🔍 常用命令:"
echo "   查看日志: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs -f"
echo "   停止服务: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down"
echo "   重启服务: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE restart"
echo "   查看状态: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps"
