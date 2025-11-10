#!/bin/bash

# 博客系统自动化部署脚本
# 使用方法: sudo ./scripts/deploy.sh [production|development]

set -e

ENV=${1:-production}
COMPOSE_FILE="docker-compose.yml"

if [ "$ENV" = "production" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
fi

echo "🚀 开始部署博客系统 (环境: $ENV)..."

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，请先安装Docker"
    exit 1
fi

# 检查Docker Compose
if docker compose version &> /dev/null; then
    COMPOSE_CMD="docker compose"
elif command -v docker-compose &> /dev/null; then
    COMPOSE_CMD="docker-compose"
else
    echo "❌ Docker Compose未安装"
    exit 1
fi

# 配置Docker日志（可选）
mkdir -p /etc/docker
if [ ! -f /etc/docker/daemon.json ]; then
    cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF
    systemctl daemon-reload
    systemctl restart docker || true
fi

# 检查必需的Docker镜像是否存在
echo "🔍 检查必需的Docker镜像..."
REQUIRED_IMAGES=(
    "docker.1ms.run/library/golang:latest"
    "docker.1ms.run/library/mysql:8.0.44"
    "docker.1ms.run/library/nginx:latest"
)

MISSING_IMAGES=()
for image in "${REQUIRED_IMAGES[@]}"; do
    if docker images "$image" --format "{{.Repository}}:{{.Tag}}" 2>/dev/null | grep -q "^${image}$" || \
       docker images --format "{{.Repository}}:{{.Tag}}" 2>/dev/null | grep -q "^${image}$" || \
       docker images --format "{{.ID}}" "$image" 2>/dev/null | grep -q .; then
        echo "✅ 镜像存在: $image"
    else
        echo "❌ 镜像不存在: $image"
        MISSING_IMAGES+=("$image")
    fi
done

if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    echo ""
    echo "❌ 以下必需的镜像不存在，请先加载镜像："
    for img in "${MISSING_IMAGES[@]}"; do
        echo "   - $img"
    done
    echo ""
    echo "💡 提示: 如果已有镜像包，请先加载："
    echo "   docker load -i <镜像包路径>"
    exit 1
fi

# 检查.env文件
if [ ! -f .env ]; then
    if [ -f env.template ]; then
        cp env.template .env
        echo "⚠️  已创建.env文件，请编辑配置后重新运行部署脚本"
        exit 1
    else
        echo "❌ 找不到.env文件或env.template模板"
        exit 1
    fi
fi

# 停止旧容器
echo "🛑 停止旧容器..."
$COMPOSE_CMD -f $COMPOSE_FILE down || true

# 构建镜像（使用本地基础镜像）
echo "🔨 构建Docker镜像..."
$COMPOSE_CMD -f $COMPOSE_FILE build --no-cache --pull=false

# 启动服务
echo "🚀 启动服务..."
$COMPOSE_CMD -f $COMPOSE_FILE up -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 15

# 检查服务状态
echo "📊 服务状态:"
$COMPOSE_CMD -f $COMPOSE_FILE ps

echo ""
echo "✅ 部署完成！"
echo ""
echo "📝 常用命令:"
echo "   查看日志: $COMPOSE_CMD -f $COMPOSE_FILE logs -f"
echo "   停止服务: $COMPOSE_CMD -f $COMPOSE_FILE down"
echo "   重启服务: $COMPOSE_CMD -f $COMPOSE_FILE restart"
