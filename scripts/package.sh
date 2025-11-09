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
    
    # 拉取镜像
    docker pull "$image" || {
        echo "⚠️  镜像拉取失败: $image"
        continue
    }
    
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
