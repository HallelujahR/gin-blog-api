#!/bin/bash

# 更新部署脚本（不停止服务，零停机更新）
set -e

ENV=${1:-production}
COMPOSE_FILE="docker-compose.yml"

if [ "$ENV" = "production" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
fi

# 检测Docker Compose命令（优先使用docker compose插件，兼容Docker 26.1+）
DOCKER_COMPOSE_CMD=""
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
else
    echo "❌ Docker Compose未安装"
    exit 1
fi

echo "✅ 使用Docker Compose命令: $DOCKER_COMPOSE_CMD"

echo "🔄 开始更新博客系统..."

# 拉取最新代码
if [ -d .git ]; then
    echo "📥 拉取最新代码..."
    git pull origin main
fi

# 重新构建镜像
echo "🔨 重新构建镜像..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE build

# 滚动更新（先更新API，再更新前端）
echo "🔄 滚动更新服务..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up -d --no-deps api
sleep 5
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up -d --no-deps frontend

# 清理旧镜像
echo "🧹 清理旧镜像..."
docker image prune -f

echo "✅ 更新完成！"

