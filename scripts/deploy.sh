#!/usr/bin/env bash

###############################################################################
# Go API 裸机部署脚本
# -----------------------------------------------------------------------------
# 1. 重新编译当前项目（GOOS=linux GOARCH=amd64）
# 2. 自动生成/更新 systemd 服务（blog-api.service）
# 3. 重载并启动服务，确保日志目录存在
#
# 用法：
#   sudo ./scripts/deploy.sh              # 默认执行 build + restart
#   sudo ./scripts/deploy.sh build        # 仅构建二进制
#   sudo ./scripts/deploy.sh restart      # 在已有二进制基础上重启服务
###############################################################################

set -euo pipefail

COMMAND="${1:-deploy}"  # deploy | build | restart
SERVICE_NAME="blog-api"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN_DIR="$PROJECT_ROOT/bin"
BINARY_PATH="$BIN_DIR/api"
LOG_DIR="$PROJECT_ROOT/logs"
ENV_FILE="$PROJECT_ROOT/.env"

# 前端与 Nginx 相关配置，可通过环境变量覆盖
FRONTEND_DIR="${FRONTEND_DIR:-/opt/blog/gin-blog-vue-font}"
FRONTEND_DIST="$FRONTEND_DIR/dist"
NGINX_CONF_FILE="${NGINX_CONF_FILE:-/etc/nginx/conf.d/blog.conf}"
NGINX_SERVER_NAME="${NGINX_SERVER_NAME:-_}"
API_INTERNAL_HOST="${API_INTERNAL_HOST:-127.0.0.1}"
API_INTERNAL_PORT="${API_INTERNAL_PORT:-8080}"
API_INTERNAL_URL="http://${API_INTERNAL_HOST}:${API_INTERNAL_PORT}"

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------
log()   { echo -e "\033[1;34m[INFO]\033[0m $*"; }
warn()  { echo -e "\033[1;33m[WARN]\033[0m $*"; }
error() { echo -e "\033[1;31m[ERR ]\033[0m $*"; }

die() {
  error "$*"
    exit 1
}

require_root() {
  if [[ $EUID -ne 0 ]]; then
    die "请使用 sudo 或 root 身份执行此脚本"
  fi
}

check_prerequisites() {
  command -v go >/dev/null 2>&1 || die "未检测到 go，请先安装 Go 1.25+"
  command -v systemctl >/dev/null 2>&1 || die "当前系统不支持 systemd"
  command -v npm >/dev/null 2>&1 || die "未检测到 npm，请先安装 Node.js 18+"
  command -v nginx >/dev/null 2>&1 || die "未检测到 nginx，请先安装 Nginx"
  [[ -f "$ENV_FILE" ]] || die "未找到 .env，请先根据 env.template 创建"
  [[ -d "$FRONTEND_DIR" ]] || die "未找到前端目录 $FRONTEND_DIR"
}

ensure_dirs() {
  mkdir -p "$BIN_DIR"
  mkdir -p "$LOG_DIR"
}

build_binary() {
  log "编译 Go 项目 (GOOS=linux GOARCH=amd64)..."
  (cd "$PROJECT_ROOT" && GOOS=linux GOARCH=amd64 go build -o "$BINARY_PATH" ./)
  log "二进制输出: $BINARY_PATH"
}

write_service_file() {
  if [[ ! -f "$SERVICE_FILE" ]]; then
    log "创建 systemd 服务文件 $SERVICE_FILE"
  else
    log "更新 systemd 服务文件 $SERVICE_FILE"
  fi
  cat <<SERVICE | tee "$SERVICE_FILE" >/dev/null
[Unit]
Description=Blog API (Go)
After=network.target mysql.service redis.service

[Service]
Type=simple
WorkingDirectory=$PROJECT_ROOT
EnvironmentFile=$ENV_FILE
ExecStart=$BINARY_PATH
Restart=on-failure
RestartSec=5
LimitNOFILE=65535
StandardOutput=append:$LOG_DIR/service.log
StandardError=append:$LOG_DIR/service.log

[Install]
WantedBy=multi-user.target
SERVICE
}

reload_service() {
  log "重载 systemd daemon"
  systemctl daemon-reload
  log "启用开机自启"
  systemctl enable "$SERVICE_NAME"
  log "重启服务"
  systemctl restart "$SERVICE_NAME"
  systemctl status "$SERVICE_NAME" --no-pager
}

build_frontend() {
  log "构建前端 ($FRONTEND_DIR)..."
  if [[ ! -d "$FRONTEND_DIR" ]]; then
    die "前端目录 $FRONTEND_DIR 不存在，无法构建"
  fi
  pushd "$FRONTEND_DIR" >/dev/null
  if [[ -f package-lock.json ]]; then
    npm ci
  else
    npm install
  fi
  npm run build
  popd >/dev/null
  if [[ ! -d "$FRONTEND_DIST" ]]; then
    die "前端构建失败，未找到 $FRONTEND_DIST"
  fi
  log "前端构建完成，输出目录 $FRONTEND_DIST"
}

write_nginx_conf() {
  log "更新 Nginx 配置 $NGINX_CONF_FILE"
  cat <<NGINX | tee "$NGINX_CONF_FILE" >/dev/null
server {
    listen 80;
    server_name $NGINX_SERVER_NAME;
    root $FRONTEND_DIST;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /api/ {
        proxy_pass $API_INTERNAL_URL;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
    }
}
NGINX

  log "检测 Nginx 配置"
  nginx -t
  log "重新加载 Nginx"
  systemctl reload nginx
}

deploy_frontend_stack() {
  build_frontend
  write_nginx_conf
}

case "$COMMAND" in
  build)
    require_root
    check_prerequisites
    ensure_dirs
    build_binary
    build_frontend
    log "仅构建完成，如需启动请执行 sudo systemctl restart $SERVICE_NAME && sudo systemctl reload nginx"
    ;;
  restart)
    require_root
    check_prerequisites
    ensure_dirs
    if [[ ! -x "$BINARY_PATH" ]]; then
      warn "未找到二进制，自动触发构建"
      build_binary
    fi
    write_service_file
    reload_service
    deploy_frontend_stack
    ;;
  deploy|*)
    require_root
    check_prerequisites
    ensure_dirs
    build_binary
    write_service_file
    reload_service
    deploy_frontend_stack
    ;;
esac

log "部署流程完成"
