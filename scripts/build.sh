#!/usr/bin/env bash
# hitest 统一构建脚本
#   - 任意平台：构建 Go 服务端二进制 build/hitest
#   - macOS + Xcode：额外构建未签名 iOS IPA build/hitest-unsigned.ipa
#
# 用法：
#   scripts/build.sh            # 能构建什么就构建什么
#   scripts/build.sh go         # 只构建 Go
#   scripts/build.sh ios        # 只构建 iOS（需 macOS）
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

GOOS_TARGET="${GOOS_TARGET:-linux}"
GOARCH_TARGET="${GOARCH_TARGET:-amd64}"
IPA_NAME="hitest-unsigned.ipa"

build_go() {
  echo "==> [Go] building build/hitest (GOOS=$GOOS_TARGET GOARCH=$GOARCH_TARGET)"
  mkdir -p build
  CGO_ENABLED=0 GOOS="$GOOS_TARGET" GOARCH="$GOARCH_TARGET" \
    go build -trimpath -ldflags '-s -w' -o build/hitest ./cmd/hitest
  if command -v upx >/dev/null 2>&1; then
    upx -q build/hitest || true
  fi
  ls -lh build/hitest
}

build_ios() {
  if [ "$(uname)" != "Darwin" ]; then
    echo "==> [iOS] 跳过：仅 macOS 可构建"
    return 0
  fi
  echo "==> [iOS] building unsigned IPA"
  cd "$ROOT/client"
  flutter config --no-enable-swift-package-manager || true
  flutter pub get
  flutter gen-l10n
  # 构建时代码生成图标（仓库不含 PNG）
  node "$ROOT/scripts/gen_icon.js"
  dart run flutter_launcher_icons
  flutter build ios --release --config-only --no-codesign
  (cd ios && pod install)
  xcodebuild \
    -workspace ios/Runner.xcworkspace \
    -scheme Runner \
    -configuration Release \
    -sdk iphoneos \
    -destination 'generic/platform=iOS' \
    -derivedDataPath build/ios \
    CODE_SIGN_STYLE=Manual \
    DEVELOPMENT_TEAM="" \
    PROVISIONING_PROFILE_SPECIFIER="" \
    CODE_SIGN_IDENTITY="" \
    CODE_SIGNING_REQUIRED=NO \
    CODE_SIGNING_ALLOWED=NO \
    build
  APP_PATH="$(find build/ios -type d -name Runner.app | head -1)"
  test -n "$APP_PATH"
  rm -rf "$ROOT/build/ipa" && mkdir -p "$ROOT/build/ipa/Payload"
  cp -R "$APP_PATH" "$ROOT/build/ipa/Payload/"
  (cd "$ROOT/build/ipa" && zip -qr "$ROOT/build/$IPA_NAME" Payload)
  ls -lh "$ROOT/build/$IPA_NAME"
}

case "${1:-all}" in
  go)  build_go ;;
  ios) build_ios ;;
  all) build_go; build_ios ;;
  *)   echo "用法: $0 [all|go|ios]"; exit 2 ;;
esac
