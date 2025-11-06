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

# 检查现有配置，如果存在则合并
if [ -f /etc/docker/daemon.json ]; then
    echo "⚠️  检测到现有配置，将合并镜像加速器配置..."
    # 使用Python合并JSON（如果可用）
    if command -v python3 &> /dev/null; then
        python3 << 'PYTHON_SCRIPT'
import json
import sys

try:
    with open('/etc/docker/daemon.json', 'r') as f:
        existing = json.load(f)
except:
    existing = {}

# 添加镜像加速器
mirrors = [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com",
    "https://registry.docker-cn.com"
]

# 如果已有镜像加速器，合并
if "registry-mirrors" in existing:
    existing_mirrors = existing["registry-mirrors"] or []
    # 合并并去重
    existing["registry-mirrors"] = list(dict.fromkeys(existing_mirrors + mirrors))
else:
    existing["registry-mirrors"] = mirrors

# 添加其他配置（如果不存在）
if "max-concurrent-downloads" not in existing:
    existing["max-concurrent-downloads"] = 10
if "log-driver" not in existing:
    existing["log-driver"] = "json-file"
if "log-opts" not in existing:
    existing["log-opts"] = {"max-size": "10m", "max-file": "3"}

# 写入新配置
with open('/etc/docker/daemon.json', 'w') as f:
    json.dump(existing, f, indent=2)
PYTHON_SCRIPT
    else
        # 如果没有Python，创建新配置（简单方式）
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
    fi
else
    # 创建新配置
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
fi

# 验证JSON格式
echo "🔍 验证配置文件格式..."
if command -v python3 &> /dev/null; then
    if ! python3 -m json.tool /etc/docker/daemon.json > /dev/null 2>&1; then
        echo "❌ JSON格式错误！正在恢复备份..."
        if [ -f /etc/docker/daemon.json.bak.* ]; then
            sudo cp /etc/docker/daemon.json.bak.* /etc/docker/daemon.json
        else
            echo "⚠️  没有备份，删除配置文件..."
            sudo rm -f /etc/docker/daemon.json
        fi
        exit 1
    fi
fi

# 重启Docker服务
echo "🔄 重启Docker服务..."
sudo systemctl daemon-reload

# 检查Docker服务状态（兼容Docker 26.1+）
if ! sudo systemctl restart docker; then
    echo "❌ Docker服务启动失败！"
    echo "📋 查看错误信息："
    sudo systemctl status docker.service --no-pager || true
    echo ""
    echo "🔧 尝试修复："
    echo "1. 检查配置文件：sudo cat /etc/docker/daemon.json"
    echo "2. 验证JSON格式：python3 -m json.tool /etc/docker/daemon.json"
    echo "3. 如果配置有问题，删除配置文件：sudo rm /etc/docker/daemon.json"
    echo "4. 然后重新运行此脚本"
    exit 1
fi

# 等待Docker服务完全启动（Docker 26.1+可能需要更多时间）
echo "⏳ 等待Docker服务启动..."
sleep 3

# 验证Docker是否正常运行
if ! docker info > /dev/null 2>&1; then
    echo "⚠️  Docker服务已启动，但可能尚未完全就绪，请稍候..."
    sleep 2
fi

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

