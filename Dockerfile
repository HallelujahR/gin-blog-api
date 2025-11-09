# 多阶段构建 - 构建阶段
FROM golang:latest AS builder

# 设置工作目录
WORKDIR /app

# 配置Go代理为阿里云镜像
ENV GOPROXY=https://mirrors.aliyun.com/goproxy/,direct
ENV GO111MODULE=on

# 安装必要的依赖
RUN apt-get update && apt-get install -y git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags '-w -s' -o api .

# 运行阶段
FROM debian:latest

# 设置工作目录
WORKDIR /app

# 安装必要的运行时依赖
RUN apt-get update && apt-get install -y ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 从构建阶段复制二进制文件
COPY --from=builder /app/api .

# 创建上传目录
RUN mkdir -p /app/uploads/images /app/uploads/files

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./api"]
