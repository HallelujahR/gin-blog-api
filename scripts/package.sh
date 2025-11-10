#!/bin/bash

# Docker镜像打包脚本
# 功能：将本地已安装的Docker镜像打包成tar文件
# 使用方法: ./scripts/package.sh

set -e

echo "📦 开始打包Docker镜像（仅使用本地镜像）..."

# 检查Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装"
    exit 1
fi

# 创建临时目录
TEMP_DIR=$(mktemp -d)
PACKAGE_DIR="$TEMP_DIR/docker-package/images"
mkdir -p "$PACKAGE_DIR"

# 定义需要打包的镜像（使用服务器上的镜像名称）
IMAGES=(
    "docker.1ms.run/library/golang:latest"
    "docker.1ms.run/library/mysql:latest"
    "docker.1ms.run/library/nginx:latest"
    "docker.1ms.run/library/node:latest"
)

# 检查并导出本地镜像
MISSING_IMAGES=()
for image in "${IMAGES[@]}"; do
    echo "🔍 检查镜像: $image"
    
    # 检查镜像是否已存在（支持多种格式）
    IMAGE_EXISTS=false
    if docker images "$image" --format "{{.Repository}}:{{.Tag}}" 2>/dev/null | grep -q "^${image}$"; then
        IMAGE_EXISTS=true
    elif docker images --format "{{.Repository}}:{{.Tag}}" 2>/dev/null | grep -q "^${image}$"; then
        IMAGE_EXISTS=true
    elif docker images --format "{{.ID}}" "$image" 2>/dev/null | grep -q .; then
        IMAGE_EXISTS=true
    fi
    
    if [ "$IMAGE_EXISTS" = true ]; then
        echo "✅ 镜像已存在: $image"
        
        # 导出镜像
        IMAGE_FILE=$(echo "$image" | tr '/:' '_').tar
        echo "💾 导出镜像: $IMAGE_FILE"
        docker save "$image" -o "$PACKAGE_DIR/$IMAGE_FILE" || {
            echo "⚠️  镜像导出失败: $image"
            continue
        }
        echo "✅ 镜像导出成功"
    else
        echo "❌ 镜像不存在: $image"
        MISSING_IMAGES+=("$image")
    fi
done

# 检查是否有缺失的镜像
if [ ${#MISSING_IMAGES[@]} -gt 0 ]; then
    echo ""
    echo "❌ 以下镜像在本地不存在，请先拉取："
    for img in "${MISSING_IMAGES[@]}"; do
        echo "   - $img"
    done
    echo ""
    echo "💡 提示: 请先运行以下命令拉取镜像："
    for img in "${MISSING_IMAGES[@]}"; do
        echo "   docker pull $img"
    done
    rm -rf "$TEMP_DIR"
    exit 1
fi

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
