# 第一阶段：构建 Golang 可执行程序
#FROM golang:1.25-alpine AS builder
#国内
FROM docker.m.daocloud.io/library/golang:1.25-alpine AS builder

WORKDIR /app

# 拷贝 go.mod 和 go.sum 预下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝全量源码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/swe-agent ./cmd/agent/main.go


# 第二阶段：构建包含完整开发工具链的运行镜像
# FROM golang:1.25-alpine
# 国内
FROM docker.m.daocloud.io/library/golang:1.25-alpine

# 安装基本软件开发工具 (bash, git, python3, curl)
RUN apk add --no-cache \
    bash \
    git \
    python3 \
    py3-pip \
    py3-requests \
    py3-bs4 \
    py3-lxml \
    curl \
    jq \
    ca-certificates \
    && rm -rf /var/cache/apk/*

WORKDIR /workspace

# 从 builder 阶段拷贝编译出的 Agent 二进制程序
COPY --from=builder /app/bin/swe-agent /usr/local/bin/swe-agent

# 设置默认启动命令
ENTRYPOINT ["swe-agent"]

