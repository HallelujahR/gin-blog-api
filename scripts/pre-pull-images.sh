#!/bin/bash

# 预先拉取Docker镜像脚本
# 用于在构建前拉取所有需要的镜像，避免构建时超时

set -e

echo "📥 预先拉取Docker镜像..."

# 检查Docker是否运行
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker未运行，请先启动Docker服务"
    echo "   执行: sudo systemctl start docker"
    exit 1
fi

# 检查镜像加速器配置
if ! docker info 2>/dev/null | grep -q "Registry Mirrors"; then
    echo "⚠️  未检测到Docker镜像加速器配置"
    echo "💡 建议先配置镜像加速器以加快拉取速度"
    echo "   执行: sudo ./scripts/configure-docker-mirror.sh"
    echo ""
    read -p "是否继续拉取镜像？(y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 定义需要拉取的镜像
IMAGES=(
    "golang:1.22"
    "mysql:8.0"
    "nginx:alpine"
    "node:latest"
)

# 拉取镜像（带超时控制，兼容Docker 26.1+）
SUCCESS=0
FAILED=0

for image in "${IMAGES[@]}"; do
    echo ""
    echo "📥 拉取镜像: $image"
    # 使用timeout命令控制超时（如果可用），默认300秒
    if command -v timeout &> /dev/null; then
        if timeout 300 docker pull "$image" 2>/dev/null || docker pull "$image"; then
            echo "✅ $image 拉取成功"
            ((SUCCESS++))
        else
            echo "❌ $image 拉取失败"
            ((FAILED++))
        fi
    else
        # 如果没有timeout命令，直接拉取
        if docker pull "$image"; then
            echo "✅ $image 拉取成功"
            ((SUCCESS++))
        else
            echo "❌ $image 拉取失败"
            ((FAILED++))
        fi
    fi
done

echo ""
echo "📊 拉取结果统计:"
echo "   ✅ 成功: $SUCCESS"
echo "   ❌ 失败: $FAILED"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "⚠️  部分镜像拉取失败"
    echo "💡 建议解决方案："
    echo "   1. 配置镜像加速器: sudo ./scripts/configure-docker-mirror.sh"
    echo "   2. 检查网络连接"
    echo "   3. 稍后重试: ./scripts/pre-pull-images.sh"
    echo ""
    echo "即使部分镜像拉取失败，构建时也会自动重试"
    exit 1
else
    echo ""
    echo "✅ 所有镜像拉取成功！可以开始部署了"
    echo "   执行: ./scripts/deploy.sh production"
fi

