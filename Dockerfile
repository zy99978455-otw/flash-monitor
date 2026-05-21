# --- Build Stage ---
FROM golang:1.24-alpine AS builder

# 安装构建依赖（如果需要 cgo 可以加 build-base，目前看 go-ethereum 可能需要）
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# 复制依赖文件并下载
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译项目（指向 cmd/api/main.go）
# CGO_ENABLED=1 是因为 go-ethereum 某些依赖可能需要 C 编译器
RUN CGO_ENABLED=1 GOOS=linux go build -o flash-monitor-api ./cmd/api

# --- Run Stage ---
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/flash-monitor-api .
# 如果你有 .env 文件或者其他静态资源，也需要复制
# COPY --from=builder /app/.env .

# 暴露端口（根据 main.go 里的配置，默认是 4010）
EXPOSE 4010

# 运行程序
CMD ["./flash-monitor-api"]
