#!/usr/bin/env bash
# hitest 构建入口。
#
# 用法:
#   scripts/build.sh go              # 构建 Go 二进制 (默认 amd64 + arm64, 含前端嵌入)
#   scripts/build.sh go --arch arm64 # 只构建指定架构
#   scripts/build.sh ios             # 构建未签名 iOS IPA (需 macOS + Xcode + Flutter)
#   scripts/build.sh all             # 两者都构建 (ios 仅在 macOS 上执行)
#   scripts/build.sh icon            # 重新生成占位 AppIcon
#
# 产物:
#   bin/hitest-{linux-amd64,linux-arm64}   # 单二进制, 三种角色 (serve/agent/client)
#   dist/hitest-ios-unsigned.ipa           # 未签名 IPA
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

TARGET="all"
ARCH="amd64,arm64"
while [[ $# -gt 0 ]]; do
  case "$1" in
    go|ios|all|icon) TARGET="$1"; shift ;;
    --arch) ARCH="$2"; shift 2 ;;
    *) echo "未知参数: $1"; exit 1 ;;
  esac
done

build_web() {
  echo "==> 构建前端 (嵌入二进制)"
  (cd web && npm install --no-audit --no-fund && npm run build)
}

build_go() {
  build_web
  echo "==> 构建 Go 二进制"
  mkdir -p bin
  for arch in "${ARCH//,/ }"; do
    echo "    linux/$arch"
    CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
      go build -ldflags="-s -w" -o "bin/hitest-linux-$arch" ./cmd/hitest
  done
  ls -la bin/
}

build_ios() {
  if [[ "$(uname)" != "Darwin" ]]; then
    echo "!! iOS 构建需要 macOS + Xcode, 当前系统 $(uname) 跳过" >&2
    exit 0
  fi
  echo "==> 构建 iOS (未签名)"
  (cd client \
    && flutter pub get \
    && flutter gen-l10n \
    && dart run flutter_launcher_icons \
    && flutter build ios --release --no-codesign)

  echo "==> 打包 IPA"
  mkdir -p dist Payload
  rm -rf Payload
  cp -r client/build/ios/iphoneos/Runner.app Payload/
  zip -qr dist/hitest-ios-unsigned.ipa Payload
  rm -rf Payload
  ls -la dist/
}

case "$TARGET" in
  go)   build_go ;;
  ios)  build_ios ;;
  icon) node scripts/gen_icon.js ;;
  all)
    build_go
    # iOS 仅在 macOS 上执行, 其他平台给出提示而非失败
    build_ios || true
    ;;
esac

echo "==> 完成"
