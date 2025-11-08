# Dockerfile.offline - 离线部署专用
# 第一阶段：使用已加载的 golang 基础镜像
FROM golang:1.25-alpine AS builder

# 设置工作目录
WORKDIR /app

# 配置阿里云Alpine镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的依赖
RUN apk add --no-cache git ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖（如果网络有问题，可以注释掉这一行，依赖应该在本地）
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api .

# 第二阶段：使用轻量级运行时镜像
FROM alpine:3.18

# 设置工作目录
WORKDIR /app

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制二进制文件
COPY --from=builder /app/api .

# 创建上传目录
RUN mkdir -p /app/uploads/images /app/uploads/files

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./api"]