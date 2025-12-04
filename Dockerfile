# -------------------------------------------------------------------
# 第一阶段：构建 (Builder)
# ✅ 修改这里：使用 1.24 版本，与你本地保持一致
# -------------------------------------------------------------------
FROM golang:1.24.11-alpine AS builder

# 设置工作目录
WORKDIR /app

# 1. 先拷贝依赖文件 (利用 Docker 缓存机制，加快构建速度)
COPY go.mod go.sum ./
# 下载依赖 (如果国内网络慢，可以在这里加 GOPROXY)
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go mod download

# 2. 拷贝所有源代码
COPY . .

# 3. 编译成二进制文件
# CGO_ENABLED=0: 关闭 C 语言依赖，确保静态链接，兼容性最好
# GOOS=linux: 目标系统是 Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o flash-monitor cmd/monitor/main.go

# -------------------------------------------------------------------
# 第二阶段：运行 (Runner)
# 使用极小的 Alpine Linux 作为基础镜像
# -------------------------------------------------------------------
FROM alpine:latest

WORKDIR /app

# 安装基础证书 (HTTPS 请求需要) 和 时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海 (Web3 常用)
ENV TZ=Asia/Shanghai

# 从第一阶段拷贝编译好的二进制文件
COPY --from=builder /app/flash-monitor .

# 创建日志和配置目录
RUN mkdir logs configs

# 拷贝示例配置 (在实际部署时，我们会用挂载的方式覆盖它)
COPY configs/config.example.yaml ./configs/config.yaml

# 声明程序运行命令
CMD ["./flash-monitor"]