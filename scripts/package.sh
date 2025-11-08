#!/bin/bash

# Docker和镜像打包脚本（本地运行）
# 功能：将Docker安装包和所需镜像打包成tar文件
# 使用方法: ./scripts/package.sh

set -e

echo "📦 开始打包Docker和镜像..."

# 创建临时目录
TEMP_DIR=$(mktemp -d)
PACKAGE_DIR="$TEMP_DIR/docker-package"
mkdir -p "$PACKAGE_DIR"

echo "📁 临时目录: $TEMP_DIR"

# ========== 打包Docker安装包（CentOS）==========
echo ""
echo "📦 准备Docker安装包..."
mkdir -p "$PACKAGE_DIR/docker-ce"

# 检查是否有Docker安装包
if [ -d "packages/docker-ce" ]; then
    echo "✅ 找到本地Docker安装包"
    cp -r packages/docker-ce/* "$PACKAGE_DIR/docker-ce/" 2>/dev/null || true
else
    echo "⚠️  未找到本地Docker安装包，请先下载Docker安装包到 packages/docker-ce/ 目录"
    echo "💡 下载命令（CentOS 7）:"
    echo "   mkdir -p packages/docker-ce"
    echo "   cd packages/docker-ce"
    echo "   yum install --downloadonly --downloaddir=. docker-ce docker-ce-cli containerd.io docker-compose-plugin"
fi

# ========== 打包Docker镜像 ==========
echo ""
echo "📦 打包Docker镜像..."

# 检查Docker是否运行
if ! command -v docker &> /dev/null; then
    echo "❌ Docker未安装，无法打包镜像"
    exit 1
fi

if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker服务未运行，请先启动Docker"
    exit 1
fi

# 定义需要打包的镜像（使用本地已有的镜像版本）
# 如果本地没有这些镜像，脚本会尝试拉取
IMAGES=(
    "golang:1.25-alpine"
    "mysql:8.0.44"
    "nginx:latest"
    "node:latest"
)

mkdir -p "$PACKAGE_DIR/images"

# 拉取或导出镜像
for image in "${IMAGES[@]}"; do
    echo "📥 处理镜像: $image"
    
    # 检查镜像是否存在
    if docker images "$image" --format "{{.Repository}}:{{.Tag}}" | grep -q "$image"; then
        echo "✅ 镜像已存在: $image"
    else
        echo "📥 镜像不存在，尝试拉取: $image"
        docker pull "$image" || {
            echo "⚠️  镜像拉取失败: $image"
            echo "💡 请确保本地有该镜像或网络连接正常"
            read -p "是否跳过此镜像？(y/n) " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                continue
            else
                exit 1
            fi
        }
    fi
    
    # 导出镜像为tar文件
    IMAGE_FILE=$(echo "$image" | tr '/:' '_')
    IMAGE_FILE="${IMAGE_FILE}.tar"
    echo "💾 导出镜像: $image -> $IMAGE_FILE"
    docker save "$image" -o "$PACKAGE_DIR/images/$IMAGE_FILE" || {
        echo "⚠️  镜像导出失败: $image"
        continue
    }
    echo "✅ 镜像导出成功: $IMAGE_FILE"
done

# ========== 创建安装脚本 ==========
echo ""
echo "📝 创建安装脚本..."
cat > "$PACKAGE_DIR/install.sh" <<'INSTALL_EOF'
#!/bin/bash

# Docker安装脚本（从tar包安装）
# 使用方法: sudo ./install.sh

set -e

if [ "$EUID" -ne 0 ]; then 
    echo "❌ 请使用sudo运行此脚本"
    exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_DIR="$SCRIPT_DIR/docker-ce"
IMAGES_DIR="$SCRIPT_DIR/images"

echo "🚀 开始安装Docker..."

# 检测操作系统
if [ -f /etc/redhat-release ]; then
    OS_TYPE="centos"
    echo "✅ 检测到CentOS系统"
elif [ -f /etc/debian_version ]; then
    OS_TYPE="debian"
    echo "✅ 检测到Debian/Ubuntu系统"
else
    echo "❌ 不支持的操作系统"
    exit 1
fi

# 安装Docker（CentOS）
if [ "$OS_TYPE" = "centos" ]; then
    # 卸载旧版本
    yum remove -y docker docker-client docker-client-latest docker-common \
        docker-latest docker-latest-logrotate docker-logrotate docker-engine 2>/dev/null || true
    
    # 安装依赖
    yum install -y yum-utils device-mapper-persistent-data lvm2
    
    # 从本地RPM包安装
    if [ -d "$DOCKER_DIR" ] && [ "$(ls -A $DOCKER_DIR/*.rpm 2>/dev/null)" ]; then
        echo "📦 从本地RPM包安装Docker..."
        yum localinstall -y $DOCKER_DIR/*.rpm
    else
        echo "❌ 未找到Docker安装包"
        exit 1
    fi
fi

# 启动Docker服务
systemctl start docker
systemctl enable docker

# 验证Docker安装
if docker --version > /dev/null 2>&1; then
    echo "✅ Docker安装成功: $(docker --version)"
else
    echo "❌ Docker安装失败"
    exit 1
fi

# 加载Docker镜像
echo ""
echo "📥 加载Docker镜像..."
if [ -d "$IMAGES_DIR" ]; then
    for image_tar in "$IMAGES_DIR"/*.tar; do
        if [ -f "$image_tar" ]; then
            echo "📥 加载镜像: $(basename $image_tar)"
            docker load -i "$image_tar" || {
                echo "⚠️  镜像加载失败: $(basename $image_tar)"
                continue
            }
        fi
    done
    echo "✅ 镜像加载完成"
else
    echo "⚠️  未找到镜像目录"
fi

echo ""
echo "✅ Docker安装完成！"
INSTALL_EOF

chmod +x "$PACKAGE_DIR/install.sh"

# ========== 创建README ==========
cat > "$PACKAGE_DIR/README.md" <<'README_EOF'
# Docker安装包说明

## 包含内容

- `docker-ce/`: Docker CE安装包（RPM文件）
- `images/`: Docker镜像tar文件
- `install.sh`: 安装脚本

## 使用方法

1. 将整个目录上传到服务器
2. 在服务器上运行安装脚本：
   ```bash
   sudo ./install.sh
   ```

## 镜像列表

- golang:1.25-alpine
- mysql:8.0.44
- nginx:latest
- node:latest

## 说明

这些镜像版本是本地已有的版本。如果本地没有这些镜像，脚本会尝试拉取。
如果拉取失败，可以手动拉取镜像后再运行打包脚本。
README_EOF

# ========== 打包成tar文件 ==========
echo ""
echo "📦 打包成tar文件..."
PACKAGE_NAME="docker-package-$(date +%Y%m%d-%H%M%S).tar.gz"
cd "$TEMP_DIR"
tar -czf "$PACKAGE_NAME" -C "$TEMP_DIR" docker-package

# 移动到当前目录
mv "$PACKAGE_NAME" "$(pwd)/"
PACKAGE_PATH="$(pwd)/$PACKAGE_NAME"

echo ""
echo "✅ 打包完成！"
echo "📦 打包文件: $PACKAGE_PATH"
echo "📊 文件大小: $(du -h "$PACKAGE_PATH" | cut -f1)"

# 清理临时目录
rm -rf "$TEMP_DIR"

echo ""
echo "📝 下一步："
echo "   1. 将 $PACKAGE_NAME 上传到服务器"
echo "   2. 在服务器上解压: tar -xzf $PACKAGE_NAME"
echo "   3. 运行安装脚本: cd docker-package && sudo ./install.sh"

