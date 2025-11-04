#!/bin/bash

# Docker镜像加速器配置脚本
# 适用于CentOS系统

set -e

echo "🔧 配置Docker镜像加速器..."

# 创建Docker配置目录
sudo mkdir -p /etc/docker

# 检查是否已有daemon.json
if [ -f /etc/docker/daemon.json ]; then
    echo "⚠️  检测到已有 /etc/docker/daemon.json"
    echo "📋 备份现有配置..."
    sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.bak.$(date +%Y%m%d_%H%M%S)
fi

# 创建或更新daemon.json
echo "📝 创建/更新Docker镜像加速器配置..."
sudo tee /etc/docker/daemon.json > /dev/null <<EOF
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com",
    "https://registry.docker-cn.com"
  ],
  "max-concurrent-downloads": 10,
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

# 重启Docker服务
echo "🔄 重启Docker服务..."
sudo systemctl daemon-reload
sudo systemctl restart docker

# 验证配置
echo "✅ 验证Docker配置..."
docker info | grep -A 10 "Registry Mirrors" || echo "⚠️  无法显示镜像加速器信息，请手动检查"

echo ""
echo "✅ Docker镜像加速器配置完成！"
echo ""
echo "📋 配置的镜像加速器："
echo "   - 中科大镜像: https://docker.mirrors.ustc.edu.cn"
echo "   - 网易镜像: https://hub-mirror.c.163.com"
echo "   - 百度镜像: https://mirror.baidubce.com"
echo "   - Docker中国: https://registry.docker-cn.com"
echo ""
echo "🔍 验证镜像加速器："
echo "   docker info | grep -A 10 'Registry Mirrors'"

