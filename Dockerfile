# 后端API Dockerfile
# 兼容旧版Docker（单阶段构建）
FROM golang:latest

# 设置工作目录
WORKDIR /app

# 安装必要的依赖
RUN apk add --no-cache git ca-certificates tzdata

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
