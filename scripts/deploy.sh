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

# 配置Docker镜像加速器（阿里云）
mkdir -p /etc/docker
cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://registry.cn-hangzhou.aliyuncs.com",
    "https://docker.mirrors.ustc.edu.cn"
  ],
  "max-concurrent-downloads": 10,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

systemctl daemon-reload
systemctl restart docker || true

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

# 构建镜像
echo "🔨 构建Docker镜像..."
$COMPOSE_CMD -f $COMPOSE_FILE build --no-cache

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
