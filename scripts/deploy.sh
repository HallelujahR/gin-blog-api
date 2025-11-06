#!/bin/bash

# 博客系统自动部署脚本
# 使用方法: ./scripts/deploy.sh [production|staging]

set -e

ENV=${1:-production}
COMPOSE_FILE="docker-compose.yml"

if [ "$ENV" = "production" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
fi

echo "🚀 开始部署博客系统 (环境: $ENV)..."
echo "📋 服务器信息:"
echo "   - 操作系统: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'"' -f2 || uname -s)"
echo "   - 内核版本: $(uname -r)"

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker版本
DOCKER_VERSION=$(docker --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
echo "📋 Docker版本: $DOCKER_VERSION"

# 检查Docker Compose是否安装
# Docker 26.1+ 使用 docker compose (插件形式)
DOCKER_COMPOSE_CMD=""
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
    echo "✅ 使用Docker Compose插件: docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
    echo "✅ 使用Docker Compose独立命令: docker-compose"
else
    echo "❌ Docker Compose未安装"
    echo "💡 Docker 26.1+ 通常包含docker compose插件，请检查安装"
    exit 1
fi

# 验证Docker Compose版本
COMPOSE_VERSION=$($DOCKER_COMPOSE_CMD version --short 2>/dev/null || echo "unknown")
echo "📋 Docker Compose版本: $COMPOSE_VERSION"

# 检查.env文件
if [ ! -f .env ]; then
    echo "⚠️  .env文件不存在，正在从.env.example创建..."
    if [ -f env.template ]; then
        cp env.template .env
    elif [ -f .env.example ]; then
        cp .env.example .env
    else
        echo "❌ 找不到环境变量模板文件"
        exit 1
    fi
    echo "📝 请编辑.env文件配置数据库和API地址"
    exit 1
fi

# 检查并配置Docker镜像加速器（如果需要）
echo "🔍 检查Docker镜像加速器配置..."
if ! docker info 2>/dev/null | grep -q "Registry Mirrors"; then
    echo "⚠️  未检测到Docker镜像加速器配置"
    echo "💡 建议配置镜像加速器以加快镜像拉取速度（特别是中国大陆服务器）"
    if [ -t 0 ]; then
        # 交互式终端
        read -p "是否现在配置镜像加速器？(y/n) " -n 1 -r
        echo
        CONFIGURE_MIRROR=$REPLY
    else
        # 非交互式（如CI/CD），默认不配置
        CONFIGURE_MIRROR="n"
    fi
    
    if [[ $CONFIGURE_MIRROR =~ ^[Yy]$ ]]; then
        if [ -f scripts/configure-docker-mirror.sh ]; then
            echo "🔧 运行镜像加速器配置脚本..."
            sudo ./scripts/configure-docker-mirror.sh
        else
            echo "⚠️  配置脚本不存在，使用手动配置..."
            sudo mkdir -p /etc/docker
            sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com"
  ]
}
EOF
            sudo systemctl daemon-reload
            sudo systemctl restart docker || echo "⚠️  Docker重启失败，请手动检查"
        fi
    else
        echo "⏭️  跳过镜像加速器配置，继续部署..."
    fi
fi

# 拉取最新代码（如果是从GitHub部署）
if [ -d .git ]; then
    echo "📥 拉取最新代码..."
    git pull origin main || echo "⚠️  Git pull失败，继续使用当前代码"
fi

# 停止旧容器
echo "🛑 停止旧容器..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down || true

# 预先拉取基础镜像（避免构建时超时）
echo "📥 预先拉取基础镜像..."
echo "💡 如果镜像拉取超时，请配置镜像加速器：sudo ./scripts/configure-docker-mirror.sh"

# 定义需要拉取的镜像
IMAGES=(
    "golang:1.22"
    "mysql:8.0"
    "nginx:alpine"
    "node:latest"
)

# 拉取镜像（带超时控制）
for image in "${IMAGES[@]}"; do
    echo "📥 拉取镜像: $image"
    if timeout 300 docker pull "$image" 2>/dev/null || docker pull "$image"; then
        echo "✅ $image 拉取成功"
    else
        echo "⚠️  $image 拉取失败，将在构建时重试"
    fi
done

# 构建镜像
echo "🔨 构建Docker镜像..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE build --no-cache

# 启动服务
echo "🚀 启动服务..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "📊 检查服务状态..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps

# 显示日志
echo "📋 最近日志:"
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs --tail=50

echo ""
echo "✅ 部署完成！"
echo ""
echo "📝 服务地址:"
echo "   - 前端: http://your-domain.com"
echo "   - API: http://your-domain.com:8080"
echo ""
echo "🔍 查看日志: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs -f"
echo "🛑 停止服务: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down"
echo "🔄 重启服务: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE restart"

