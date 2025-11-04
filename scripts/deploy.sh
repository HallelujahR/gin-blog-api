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

# 检查Docker是否安装
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose是否安装
DOCKER_COMPOSE_CMD=""
if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
elif docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    echo "❌ Docker Compose未安装，请先安装Docker Compose"
    exit 1
fi

echo "✅ 使用Docker Compose命令: $DOCKER_COMPOSE_CMD"

# 检查.env文件
if [ ! -f .env ]; then
    echo "⚠️  .env文件不存在，正在从.env.example创建..."
    cp .env.example .env
    echo "📝 请编辑.env文件配置数据库和API地址"
    exit 1
fi

# 拉取最新代码（如果是从GitHub部署）
if [ -d .git ]; then
    echo "📥 拉取最新代码..."
    git pull origin main || echo "⚠️  Git pull失败，继续使用当前代码"
fi

# 停止旧容器
echo "🛑 停止旧容器..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down || true

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

