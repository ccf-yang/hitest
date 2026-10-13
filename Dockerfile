# hitest 服务端镜像：多阶段构建
#   构建: docker build -t hitest:latest .
#   运行: docker run -d --name hitest -p 9999:9999 -v hitest-data:/opt/hitest hitest:latest
#
# 首次启动进入初始化模式，浏览器打开 http://<主机>:9999 完成初始化。

# ---------- 构建阶段 ----------
FROM golang:1.26-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags '-s -w' -o /out/hitest ./cmd/hitest

# ---------- 运行阶段 ----------
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata curl \
    && addgroup -S hitest && adduser -S hitest -G hitest || true

COPY --from=builder /out/hitest /usr/local/bin/hitest

# 基础目录（配置/数据库/日志都在 <BASE_DIR>/hitest 下）
ENV HITEST_BASE_DIR=/opt \
    TZ=Asia/Shanghai \
    GIN_MODE=release

VOLUME ["/opt/hitest"]
EXPOSE 9999

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
  CMD curl -sf http://127.0.0.1:9999/api/v2/health/check >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/hitest"]
