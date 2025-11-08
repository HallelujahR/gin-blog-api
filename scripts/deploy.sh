#!/bin/bash

# 博客系统自动部署脚本（简化版）
# 功能：从tar包加载镜像并部署应用
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

# ========== 检查Docker ==========
echo ""
echo "🔧 检查Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker服务未运行，请先启动Docker: sudo systemctl start docker"
    exit 1
fi

DOCKER_VERSION=$(docker --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
echo "✅ Docker已安装，版本: $DOCKER_VERSION"

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
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

COMPOSE_VERSION=$($DOCKER_COMPOSE_CMD version --short 2>/dev/null || echo "unknown")
echo "📋 Docker Compose版本: $COMPOSE_VERSION"

# ========== 配置Docker镜像加速器 ==========
echo ""
echo "🔧 配置Docker镜像加速器..."
mkdir -p /etc/docker

# 检查是否已有配置
if [ ! -f /etc/docker/daemon.json ] || ! grep -q "registry-mirrors" /etc/docker/daemon.json; then
    # 备份现有配置
    if [ -f /etc/docker/daemon.json ]; then
        cp /etc/docker/daemon.json /etc/docker/daemon.json.bak.$(date +%Y%m%d_%H%M%S)
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
    
    # 重启Docker服务
    systemctl daemon-reload
    systemctl restart docker
    sleep 3
    echo "✅ Docker镜像加速器配置完成"
else
    echo "✅ Docker镜像加速器已配置"
fi

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

# ========== 从tar包加载镜像 ==========
echo ""
echo "📥 从tar包加载Docker镜像..."

# 检查tar包是否存在（项目根目录）
if [ ! -f "$DOCKER_PACKAGE" ]; then
    echo "❌ Docker镜像包不存在: $DOCKER_PACKAGE"
    echo "💡 请确保tar包在项目根目录，或指定正确的路径"
    exit 1
fi

# 解压tar包到临时目录
EXTRACT_DIR=$(mktemp -d)
echo "📦 解压镜像包..."
tar -xzf "$DOCKER_PACKAGE" -C "$EXTRACT_DIR"

IMAGES_DIR="$EXTRACT_DIR/docker-package/images"
if [ ! -d "$IMAGES_DIR" ]; then
    echo "❌ 镜像目录不存在: $IMAGES_DIR"
    rm -rf "$EXTRACT_DIR"
    exit 1
fi

# 加载所有镜像
echo "📥 加载Docker镜像..."
LOADED=0
FAILED=0

for image_tar in "$IMAGES_DIR"/*.tar; do
    if [ -f "$image_tar" ]; then
        IMAGE_NAME=$(basename "$image_tar" .tar)
        echo "📥 加载镜像: $IMAGE_NAME"
        if docker load -i "$image_tar" 2>&1; then
            echo "✅ $IMAGE_NAME 加载成功"
            ((LOADED++))
        else
            echo "⚠️  $IMAGE_NAME 加载失败"
            ((FAILED++))
        fi
    fi
done

# 清理临时目录
rm -rf "$EXTRACT_DIR"

if [ $LOADED -gt 0 ]; then
    echo "✅ 成功加载 $LOADED 个镜像"
fi
if [ $FAILED -gt 0 ]; then
    echo "⚠️  $FAILED 个镜像加载失败"
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
