#!/usr/bin/env bash
# hitest 容器封装脚本
#   scripts/docker.sh build            构建镜像 hitest:latest
#   scripts/docker.sh run [TAG]        启动容器（数据卷 hitest-data 持久化）
#   scripts/docker.sh stop             停止并移除容器
#   scripts/docker.sh logs             跟踪日志
#   scripts/docker.sh restart [TAG]    重启容器
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
IMAGE="${HITEST_IMAGE:-hitest:latest}"
NAME="${HITEST_CONTAINER:-hitest}"
PORT="${HITEST_PORT:-9999}"

action="${1:-help}"
tag="${2:-}"

case "$action" in
  build)
    echo "==> 构建镜像 $IMAGE"
    docker build -t "$IMAGE" "$ROOT"
    ;;
  run)
    echo "==> 启动 $NAME ($IMAGE)，端口 $PORT，数据卷 hitest-data"
    docker run -d --name "$NAME" \
      -p "$PORT":9999 \
      -v hitest-data:/opt/hitest \
      --restart unless-stopped \
      "$IMAGE"
    echo "==> 初始化: 浏览器打开 http://<主机>:$PORT"
    ;;
  stop)
    docker rm -f "$NAME" 2>/dev/null && echo "==> 已停止并移除 $NAME" || echo "==> 容器不存在"
    ;;
  logs)
    docker logs -f "$NAME"
    ;;
  restart)
    docker rm -f "$NAME" 2>/dev/null || true
    "$0" run "${tag:-$IMAGE}"
    ;;
  *)
    echo "用法: $0 {build|run [tag]|stop|logs|restart [tag]}"
    exit 2
    ;;
esac
