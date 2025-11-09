#!/bin/bash

# Docker镜像打包脚本
# 功能：将项目所需的Docker镜像打包成tar文件
# 使用方法: ./scripts/package.sh

set -e

echo "📦 开始打包Docker镜像..."

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装"
    exit 1
fi

# 配置Docker镜像加速器（Linux系统，需要root权限）
echo "🔧 检查Docker镜像加速器配置..."
if [ "$(uname)" = "Linux" ]; then
    if [ -w /etc/docker/daemon.json ] || [ "$EUID" -eq 0 ]; then
        mkdir -p /etc/docker
        if [ ! -f /etc/docker/daemon.json ] || ! grep -q "registry-mirrors" /etc/docker/daemon.json 2>/dev/null; then
            cat > /etc/docker/daemon.json <<EOF
{
  "registry-mirrors": [
    "https://docker.1ms.run/library"
  ],
  "max-concurrent-downloads": 10
}
EOF
            echo "✅ 已配置Docker镜像加速器"
            if command -v systemctl &> /dev/null && systemctl is-active --quiet docker 2>/dev/null; then
                echo "🔄 重启Docker服务..."
                systemctl daemon-reload
                systemctl restart docker || true
                sleep 3
            fi
        else
            echo "✅ Docker镜像加速器已配置"
        fi
    else
        echo "⚠️  需要root权限配置镜像加速器，跳过配置"
        echo "💡 如果拉取失败，请手动配置: sudo vi /etc/docker/daemon.json"
    fi
else
    echo "ℹ️  macOS系统检测到"
    echo "💡 如需配置镜像加速器，请："
    echo "   1. 打开 Docker Desktop"
    echo "   2. 进入 Settings > Docker Engine"
    echo "   3. 添加以下配置："
    echo '      "registry-mirrors": ['
    echo '        "https://docker.1ms.run/library"'
    echo '      ]'
    echo "   4. 点击 Apply & Restart"
fi

# 创建临时目录
TEMP_DIR=$(mktemp -d)
PACKAGE_DIR="$TEMP_DIR/docker-package/images"
mkdir -p "$PACKAGE_DIR"

# 定义需要打包的镜像（通过阿里云镜像加速器拉取）
IMAGES=(
    "golang:latest"
    "mysql:8.0"
    "nginx:latest"
    "debian:latest"
)

# 拉取并导出镜像
for image in "${IMAGES[@]}"; do
    echo "📥 处理镜像: $image"
    
    # 检查镜像是否已存在
    if docker images "$image" --format "{{.Repository}}:{{.Tag}}" 2>/dev/null | grep -q "^${image}$"; then
        echo "✅ 镜像已存在: $image"
    else
        # 拉取镜像（最多重试3次）
        RETRY_COUNT=0
        MAX_RETRIES=3
        PULL_SUCCESS=false
        while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
            if docker pull "$image" 2>&1; then
                echo "✅ 镜像拉取成功: $image"
                PULL_SUCCESS=true
                break
            else
                RETRY_COUNT=$((RETRY_COUNT + 1))
                if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
                    echo "⚠️  镜像拉取失败，重试中 ($RETRY_COUNT/$MAX_RETRIES)..."
                    sleep 2
                else
                    echo "❌ 镜像拉取失败: $image (已重试 $MAX_RETRIES 次)"
                    echo "💡 提示: 请检查网络连接或手动配置Docker镜像加速器"
                    PULL_SUCCESS=false
                fi
            fi
        done
        
        if [ "$PULL_SUCCESS" = false ]; then
            echo "⏭️  跳过镜像: $image"
            continue
        fi
    fi
    
    # 导出镜像
    IMAGE_FILE=$(echo "$image" | tr '/:' '_').tar
    echo "💾 导出镜像: $IMAGE_FILE"
    docker save "$image" -o "$PACKAGE_DIR/$IMAGE_FILE" || {
        echo "⚠️  镜像导出失败: $image"
        continue
    }
    echo "✅ 镜像导出成功"
done

# 打包成tar文件
echo "📦 打包成tar文件..."
PACKAGE_NAME="docker-images.tar.gz"
CURRENT_DIR=$(pwd)
cd "$TEMP_DIR"
tar -czf "$PACKAGE_NAME" docker-package
mv "$PACKAGE_NAME" "$CURRENT_DIR/$PACKAGE_NAME"
cd "$CURRENT_DIR"

# 清理临时目录
rm -rf "$TEMP_DIR"

echo ""
echo "✅ 打包完成！"
echo "📦 文件位置: $CURRENT_DIR/$PACKAGE_NAME"
echo "📊 文件大小: $(du -h "$CURRENT_DIR/$PACKAGE_NAME" | cut -f1)"
