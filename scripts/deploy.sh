#!/bin/bash

# 博客系统自动部署脚本（CentOS）
# 功能：安装Docker、配置镜像源、部署应用
# 使用方法: sudo ./scripts/deploy.sh [production|staging]

set -e

ENV=${1:-production}
COMPOSE_FILE="docker-compose.yml"

if [ "$ENV" = "production" ]; then
    COMPOSE_FILE="docker-compose.prod.yml"
fi

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用sudo运行此脚本"
    exit 1
fi

echo "🚀 开始部署博客系统 (环境: $ENV)..."
echo "📋 服务器信息:"
echo "   - 操作系统: $(cat /etc/os-release 2>/dev/null | grep PRETTY_NAME | cut -d'"' -f2 || uname -s)"
echo "   - 内核版本: $(uname -r)"

# 检测操作系统
if [ -f /etc/redhat-release ]; then
    OS_TYPE="centos"
    echo "✅ 检测到CentOS系统"
elif [ -f /etc/debian_version ]; then
    OS_TYPE="debian"
    echo "✅ 检测到Debian/Ubuntu系统"
else
    echo "⚠️  未知操作系统，假设为CentOS"
    OS_TYPE="centos"
fi

# ========== 配置yum阿里云镜像源（CentOS）==========
if [ "$OS_TYPE" = "centos" ]; then
    echo ""
    echo "🔧 配置yum阿里云镜像源..."
    
    # 备份原有yum源
    if [ ! -f /etc/yum.repos.d/CentOS-Base.repo.bak ]; then
        cp -a /etc/yum.repos.d/CentOS-Base.repo /etc/yum.repos.d/CentOS-Base.repo.bak 2>/dev/null || true
        echo "✅ 已备份原有yum源"
    fi
    
    # 检测CentOS版本
    CENTOS_VERSION=$(cat /etc/redhat-release | grep -oE '[0-9]+' | head -1)
    
    if [ "$CENTOS_VERSION" = "7" ]; then
        echo "📋 检测到CentOS 7，配置阿里云镜像源..."
        cat > /etc/yum.repos.d/CentOS-Base.repo <<'EOF'
[base]
name=CentOS-$releasever - Base - mirrors.aliyun.com
failovermethod=priority
baseurl=http://mirrors.aliyun.com/centos/$releasever/os/$basearch/
gpgcheck=1
gpgkey=http://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7

[updates]
name=CentOS-$releasever - Updates - mirrors.aliyun.com
failovermethod=priority
baseurl=http://mirrors.aliyun.com/centos/$releasever/updates/$basearch/
gpgcheck=1
gpgkey=http://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7

[extras]
name=CentOS-$releasever - Extras - mirrors.aliyun.com
failovermethod=priority
baseurl=http://mirrors.aliyun.com/centos/$releasever/extras/$basearch/
gpgcheck=1
gpgkey=http://mirrors.aliyun.com/centos/RPM-GPG-KEY-CentOS-7
EOF
    elif [ "$CENTOS_VERSION" = "8" ] || [ "$CENTOS_VERSION" = "9" ]; then
        echo "📋 检测到CentOS $CENTOS_VERSION，配置阿里云镜像源..."
        sed -e 's|^mirrorlist=|#mirrorlist=|g' \
            -e 's|^#baseurl=http://mirror.centos.org|baseurl=https://mirrors.aliyun.com|g' \
            -i /etc/yum.repos.d/CentOS-*.repo 2>/dev/null || true
    fi
    
    # 清理yum缓存
    yum clean all
    yum makecache
    echo "✅ yum阿里云镜像源配置完成"
fi

# ========== 修复Docker仓库配置（如果存在问题）==========
if [ "$OS_TYPE" = "centos" ] && [ -f /etc/yum.repos.d/docker-ce.repo ]; then
    echo ""
    echo "🔍 检查Docker仓库配置..."
    # 检查是否包含官方源URL
    if grep -q "download.docker.com" /etc/yum.repos.d/docker-ce.repo; then
        echo "⚠️  检测到Docker仓库使用官方源，正在修复为阿里云镜像..."
        # 备份原配置
        cp /etc/yum.repos.d/docker-ce.repo /etc/yum.repos.d/docker-ce.repo.bak.$(date +%Y%m%d_%H%M%S)
        # 替换为阿里云镜像
        sed -i 's|https://download.docker.com|https://mirrors.aliyun.com/docker-ce|g' /etc/yum.repos.d/docker-ce.repo
        sed -i 's|http://download.docker.com|https://mirrors.aliyun.com/docker-ce|g' /etc/yum.repos.d/docker-ce.repo
        echo "✅ Docker仓库配置已修复"
        # 清理yum缓存
        yum clean all
        yum makecache fast
    fi
fi

# ========== 安装Docker ==========
echo ""
echo "🔧 检查Docker安装状态..."
if ! command -v docker &> /dev/null; then
    echo "📦 安装Docker..."
    
    if [ "$OS_TYPE" = "centos" ]; then
        # 卸载旧版本Docker
        yum remove -y docker docker-client docker-client-latest docker-common \
            docker-latest docker-latest-logrotate docker-logrotate docker-engine 2>/dev/null || true
        
        # 安装依赖
        yum install -y yum-utils device-mapper-persistent-data lvm2
        
        # 删除可能存在的旧Docker仓库配置（使用官方源）
        if [ -f /etc/yum.repos.d/docker-ce.repo ]; then
            if grep -q "download.docker.com" /etc/yum.repos.d/docker-ce.repo; then
                echo "🗑️  删除使用官方源的旧Docker仓库配置..."
                rm -f /etc/yum.repos.d/docker-ce.repo
            fi
        fi
        
        # 如果Docker仓库配置不存在，创建阿里云镜像配置
        if [ ! -f /etc/yum.repos.d/docker-ce.repo ]; then
            echo "📝 创建Docker阿里云仓库配置..."
            cat > /etc/yum.repos.d/docker-ce.repo <<'EOF'
[docker-ce-stable]
name=Docker CE Stable - $basearch
baseurl=https://mirrors.aliyun.com/docker-ce/linux/centos/$releasever/$basearch/stable
enabled=1
gpgcheck=1
gpgkey=https://mirrors.aliyun.com/docker-ce/linux/centos/gpg
EOF
        else
            # 确保使用阿里云镜像
            sed -i 's|https://download.docker.com|https://mirrors.aliyun.com/docker-ce|g' /etc/yum.repos.d/docker-ce.repo
            sed -i 's|http://download.docker.com|https://mirrors.aliyun.com/docker-ce|g' /etc/yum.repos.d/docker-ce.repo
        fi
        
        # 清理yum缓存
        yum clean all
        yum makecache fast
        
        # 安装Docker
        echo "📦 正在安装Docker（使用阿里云镜像源）..."
        yum install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        
    elif [ "$OS_TYPE" = "debian" ]; then
        # 卸载旧版本
        apt-get remove -y docker docker-engine docker.io containerd runc 2>/dev/null || true
        
        # 安装依赖
        apt-get update
        apt-get install -y apt-transport-https ca-certificates curl gnupg lsb-release
        
        # 添加Docker阿里云GPG密钥
        curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/debian/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        
        # 添加Docker阿里云仓库
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://mirrors.aliyun.com/docker-ce/linux/debian $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
        
        # 安装Docker
        apt-get update
        apt-get install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
    fi
    
    # 启动Docker服务
    systemctl start docker
    systemctl enable docker
    echo "✅ Docker安装完成"
else
    DOCKER_VERSION=$(docker --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+' | head -1)
    echo "✅ Docker已安装，版本: $DOCKER_VERSION"
fi

# ========== 配置Docker镜像加速器 ==========
echo ""
echo "🔧 配置Docker镜像加速器..."
mkdir -p /etc/docker

# 检查是否已有配置
if [ -f /etc/docker/daemon.json ]; then
    # 备份现有配置
    cp /etc/docker/daemon.json /etc/docker/daemon.json.bak.$(date +%Y%m%d_%H%M%S)
fi

# 创建或更新daemon.json
cat > /etc/docker/daemon.json <<'EOF'
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
systemctl daemon-reload
systemctl restart docker
sleep 3
echo "✅ Docker镜像加速器配置完成"

# ========== 检测Docker Compose ==========
echo ""
echo "🔧 检查Docker Compose..."
DOCKER_COMPOSE_CMD=""
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
    echo "✅ 使用Docker Compose插件: docker compose"
elif command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
    echo "✅ 使用Docker Compose独立命令: docker-compose"
else
    echo "⚠️  Docker Compose未安装，尝试安装..."
    if [ "$OS_TYPE" = "centos" ]; then
        # Docker 26.1+ 通常包含docker compose插件
        # 如果没有，安装docker-compose独立版本
        curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
        chmod +x /usr/local/bin/docker-compose
        DOCKER_COMPOSE_CMD="docker-compose"
    fi
fi

COMPOSE_VERSION=$($DOCKER_COMPOSE_CMD version --short 2>/dev/null || echo "unknown")
echo "📋 Docker Compose版本: $COMPOSE_VERSION"

# ========== 检查.env文件 ==========
echo ""
echo "🔍 检查环境配置文件..."
if [ ! -f .env ]; then
    echo "⚠️  .env文件不存在，正在从模板创建..."
    if [ -f env.template ]; then
        cp env.template .env
        echo "✅ 已创建.env文件，请编辑配置数据库和API地址"
        echo "   编辑命令: vi .env"
        exit 1
    elif [ -f .env.example ]; then
        cp .env.example .env
        echo "✅ 已创建.env文件，请编辑配置数据库和API地址"
        echo "   编辑命令: vi .env"
        exit 1
    else
        echo "❌ 找不到环境变量模板文件"
        exit 1
    fi
fi

# ========== 拉取最新代码 ==========
if [ -d .git ]; then
    echo ""
    echo "📥 拉取最新代码..."
    git pull origin main || echo "⚠️  Git pull失败，继续使用当前代码"
fi

# ========== 预先拉取基础镜像 ==========
echo ""
echo "📥 预先拉取基础镜像（使用golang:1.22）..."
IMAGES=(
    "golang:1.22"
    "mysql:8.0"
    "nginx:alpine"
    "node:latest"
)

for image in "${IMAGES[@]}"; do
    echo "📥 拉取镜像: $image"
    if timeout 300 docker pull "$image" 2>/dev/null || docker pull "$image"; then
        echo "✅ $image 拉取成功"
    else
        echo "⚠️  $image 拉取失败，将在构建时重试"
    fi
done

# ========== 停止旧容器 ==========
echo ""
echo "🛑 停止旧容器..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down || true

# ========== 构建镜像 ==========
echo ""
echo "🔨 构建Docker镜像..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE build --no-cache

# ========== 启动服务 ==========
echo ""
echo "🚀 启动服务..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE up -d

# ========== 等待服务启动 ==========
echo ""
echo "⏳ 等待服务启动..."
sleep 10

# ========== 检查服务状态 ==========
echo ""
echo "📊 检查服务状态..."
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps

# ========== 显示日志 ==========
echo ""
echo "📋 最近日志:"
$DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs --tail=50

echo ""
echo "✅ 部署完成！"
echo ""
echo "📝 服务地址:"
echo "   - 前端: http://your-domain.com"
echo "   - API: http://your-domain.com:8080"
echo ""
echo "🔍 常用命令:"
echo "   查看日志: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE logs -f"
echo "   停止服务: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE down"
echo "   重启服务: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE restart"
echo "   查看状态: $DOCKER_COMPOSE_CMD -f $COMPOSE_FILE ps"
