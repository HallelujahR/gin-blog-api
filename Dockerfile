# 后端API Dockerfile
# 兼容旧版Docker（单阶段构建）
FROM golang:latest

# 设置工作目录
WORKDIR /app

# 配置阿里云apt镜像源（加速包下载）
RUN sed -i 's/deb.debian.org/mirrors.aliyun.com/g' /etc/apt/sources.list.d/debian.sources 2>/dev/null || \
    sed -i 's|http://deb.debian.org|http://mirrors.aliyun.com|g' /etc/apt/sources.list 2>/dev/null || \
    echo "deb http://mirrors.aliyun.com/debian/ bookworm main" > /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian/ bookworm-updates main" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/debian-security/ bookworm-security main" >> /etc/apt/sources.list

# 安装必要的依赖（Debian基础镜像使用apt）
RUN apt-get update && \
    apt-get install -y --no-install-recommends git ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*

# 设置时区
ENV TZ=Asia/Shanghai

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api .

# 创建上传目录
RUN mkdir -p /app/uploads/images /app/uploads/files

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./api"]
