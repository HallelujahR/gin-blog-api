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

# 构建前端（若存在指定目录）
FRONT_DIR="/opt/blog/gin-blog-vue-font"
if [ -d "$FRONT_DIR" ]; then
    echo "🧱 构建前端(Vue)..."
    docker run --rm \
      -v "$FRONT_DIR":/app \
      -w /app \
      docker.1ms.run/library/node:latest sh -lc "corepack enable || true; (npm ci || npm install) && npm run build"
    # 简单校验
    if [ ! -d "$FRONT_DIR/dist" ]; then
        echo "❌ 前端构建失败：未找到 $FRONT_DIR/dist"
        exit 1
    fi
else
    echo "ℹ️ 未检测到前端目录 $FRONT_DIR，跳过前端构建"
fi

# 确保存在默认的Nginx配置（用于 /api 反向代理）
NGINX_CONF_DIR="./docker/nginx/conf.d"
mkdir -p "$NGINX_CONF_DIR"
DEFAULT_CONF="$NGINX_CONF_DIR/default.conf"
if [ ! -f "$DEFAULT_CONF" ]; then
cat > "$DEFAULT_CONF" <<'CONF'
server {
    listen 80;
    server_name _;

    # 前端静态资源
    location / {
        root /usr/share/nginx/html;
        try_files $uri $uri/ /index.html;
    }

    # 后端API代理
    location /api/ {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
CONF
fi

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
