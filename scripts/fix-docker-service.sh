#!/bin/bash

# Docker服务修复脚本
# 用于修复Docker服务启动失败问题

set -e

echo "🔧 修复Docker服务..."

# 检查Docker服务状态
echo "📋 检查Docker服务状态..."
sudo systemctl status docker.service || true

# 检查配置文件
echo ""
echo "📋 检查Docker配置文件..."
if [ -f /etc/docker/daemon.json ]; then
    echo "发现配置文件: /etc/docker/daemon.json"
    echo "配置文件内容："
    sudo cat /etc/docker/daemon.json
    
    # 验证JSON格式
    if command -v python3 &> /dev/null; then
        echo ""
        echo "🔍 验证JSON格式..."
        if python3 -m json.tool /etc/docker/daemon.json > /dev/null 2>&1; then
            echo "✅ JSON格式正确"
        else
            echo "❌ JSON格式错误！"
            echo "修复方案："
            echo "1. 备份当前配置：sudo cp /etc/docker/daemon.json /etc/docker/daemon.json.broken"
            echo "2. 删除配置文件：sudo rm /etc/docker/daemon.json"
            echo "3. 重新运行配置脚本：sudo ./scripts/configure-docker-mirror.sh"
            exit 1
        fi
    fi
else
    echo "未找到配置文件，这是正常的"
fi

# 尝试停止Docker服务
echo ""
echo "🛑 停止Docker服务..."
sudo systemctl stop docker.service || true

# 清理可能的残留进程
echo "🧹 清理残留进程..."
sudo pkill -f dockerd || true
sleep 2

# 重新加载systemd
echo "🔄 重新加载systemd配置..."
sudo systemctl daemon-reload

# 尝试启动Docker服务
echo "🚀 启动Docker服务..."
if sudo systemctl start docker.service; then
    echo "✅ Docker服务启动成功！"
    sleep 2
    sudo systemctl status docker.service
else
    echo "❌ Docker服务启动失败！"
    echo ""
    echo "📋 详细错误信息："
    sudo journalctl -xe -u docker.service --no-pager | tail -30
    echo ""
    echo "🔧 修复建议："
    echo "1. 如果配置文件有问题，删除它："
    echo "   sudo rm /etc/docker/daemon.json"
    echo "   sudo systemctl restart docker"
    echo ""
    echo "2. 检查Docker日志："
    echo "   sudo journalctl -u docker.service -n 50"
    echo ""
    echo "3. 检查系统资源："
    echo "   free -h"
    echo "   df -h"
    echo ""
    echo "4. 重新安装Docker（最后手段）："
    echo "   参考 DEPLOYMENT_CENTOS.md 中的安装步骤"
    exit 1
fi

echo ""
echo "✅ Docker服务修复完成！"
echo "🔍 验证Docker功能："
docker --version
docker ps

