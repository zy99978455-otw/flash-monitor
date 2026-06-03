# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译项目（指向 cmd/api/main.go）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o flash-monitor-api ./cmd/api

# --- Run Stage ---
FROM alpine:latest

# 安装根证书（为了连通 Infura HTTPS）和时区数据
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/flash-monitor-api .

# 暴露端口（根据 main.go 里的配置，默认是 4010）
EXPOSE 4010

# 运行程序
CMD ["./flash-monitor-api"]
