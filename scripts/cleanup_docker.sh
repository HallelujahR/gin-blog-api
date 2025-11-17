#!/usr/bin/env bash

###############################################################################
# Docker Cleanup Utility
# ---------------------------------------------------------------------------
# 释放磁盘空间，清理以下资源：
#   1. 已停止的容器
#   2. 悬空镜像（<none>）
#   3. 不再引用的镜像与缓存
#   4. 未使用的卷与构建缓存
#
# 使用前请确认当前机器允许删除上述 Docker 资源。
# 可通过 DOCKER_PRUNE_KEEP_BUILD=1 来保留 Builder 缓存。
###############################################################################

set -euo pipefail

info()  { echo "ℹ️  $*"; }
warn()  { echo "⚠️  $*"; }
ok()    { echo "✅ $*"; }

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

info "当前磁盘占用："
df -h .

warn "将要删除所有已停止容器、悬空镜像、未使用卷与构建缓存。"

info "🧹 删除已停止的容器..."
docker container prune -f || true

info "🧹 删除悬空镜像(<none>)..."
dangling_images=$(docker images -f "dangling=true" -q | sort -u || true)
if [ -n "${dangling_images:-}" ]; then
  docker rmi $dangling_images || true
else
  info "没有悬空镜像需要清理"
fi

info "🧹 删除未使用的网络..."
docker network prune -f || true

info "🧹 删除未使用的数据卷..."
docker volume prune -f || true

info "🧹 删除未使用的镜像与容器数据..."
docker system prune -f || true

if [ "${DOCKER_PRUNE_KEEP_BUILD:-0}" != "1" ]; then
  info "🧹 删除构建缓存..."
  docker builder prune -a -f || true
else
  info "跳过构建缓存清理 (DOCKER_PRUNE_KEEP_BUILD=1)"
fi

info "📦 当前 Docker 镜像占用："
docker system df

ok "Docker 清理完成，可重新执行部署脚本。"

