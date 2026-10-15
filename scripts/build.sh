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
  IFS=',' read -ra __ARCHS <<< "$ARCH"
  for arch in "${__ARCHS[@]}"; do
    echo "    linux/$arch"
    CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
      go build -ldflags="-s -w" -o "bin/hitest-linux-$arch" ./cmd/hitest
  done
  ls -la bin/
}

# bundle_executable <Info.plist>: 取 CFBundleExecutable。
# 真包的 Info.plist 是二进制 plist(bplist00), XML 解析读不到 → 有 plutil 且是二进制就走 plutil。
bundle_executable() {
  local P="$1" MAGIC
  # 都带 || true: 脚本顶层是 set -euo pipefail, 赋值里 $() 失败会直接终止整个构建，
  # 这里要的是"读不到就空着, 由调用方报错"
  MAGIC=$(head -c 6 "$P" 2>/dev/null || true)
  case "$MAGIC" in
    bplist*) { command -v plutil >/dev/null 2>&1 && plutil -extract CFBundleExecutable raw -o - "$P" 2>/dev/null; } || true ;;
    *) { tr '\n\r\t' '   ' < "$P" | sed 's/<[^>]*>/ /g' \
          | awk '{ for (i = 1; i < NF; i++) if ($i == "CFBundleExecutable") { print $(i + 1); exit } }'; } || true ;;
  esac
}

# package_ipa <.app 目录> <产物 ipa 路径>
# 安装器靠 Payload/ 下那层 <App>.app 目录定位 bundle, 所以打包前后各查一道。
package_ipa() {
  local APP="$1" OUT="$2" NAME OUT_DIR OUT_DIR_ABS STAGE EXE LIST
  if [ ! -d "$APP" ]; then echo "!! 没找到 $APP" >&2; return 1; fi
  NAME=$(basename "$APP")
  # bundle 目录名必须以 .app 结尾, 这是安装器唯一的定位依据。
  # 老 bug 就是 `cp -r Runner.app Payload/` 在 Payload 不存在时把 Payload 变成了 bundle,
  # 归档里成了 Payload/Info.plist —— 那种包这里直接拒绝。
  case "$NAME" in
    *.app) ;;
    *) echo "!! $APP 不是 .app bundle（目录名要以 .app 结尾）, 这样打出来装不上" >&2; return 1 ;;
  esac

  # zip 在暂存目录里执行, 产物路径必须先变绝对（先建目录再取绝对路径）。
  # || true 是为了: 顶层 set -euo pipefail 下, 赋值里的 $() 失败会直接终止整个构建,
  # 而这里希望走下面的显式报错。
  OUT_DIR=$(dirname "$OUT")
  mkdir -p "$OUT_DIR"
  OUT_DIR_ABS=$( (cd "$OUT_DIR" && pwd) || true)
  if [ -z "$OUT_DIR_ABS" ]; then
    echo "!! 产物目录解析失败: $OUT_DIR" >&2
    return 1
  fi
  OUT="$OUT_DIR_ABS/$(basename "$OUT")"
  rm -f "$OUT"

  # 第一道（文件系统层）: 别在仓库根建 Payload —— 顺序写反会把 .app 摊平成 Payload/Info.plist
  STAGE=$(mktemp -d)
  mkdir -p "$STAGE/Payload"
  cp -R "$APP" "$STAGE/Payload/"
  if [ ! -f "$STAGE/Payload/$NAME/Info.plist" ]; then
    echo "!! 暂存后没有 Payload/$NAME/Info.plist, .app 没进去（检查 $APP）" >&2
    rm -rf "$STAGE"; return 1
  fi
  EXE=$(bundle_executable "$STAGE/Payload/$NAME/Info.plist" || true)
  # 读不到 CFBundleExecutable 就算失败: 正常的 flutter bundle 一定有这一项,
  # 空值只可能是 Info.plist 损坏/被截断 —— 不能因为"不知道叫什么"就放行
  if [ -z "$EXE" ]; then
    echo "!! 读不到 CFBundleExecutable, $NAME/Info.plist 可能损坏" >&2
    rm -rf "$STAGE"; return 1
  fi
  if [ ! -f "$STAGE/Payload/$NAME/$EXE" ]; then
    echo "!! Payload/$NAME 里没有可执行文件 $EXE" >&2
    rm -rf "$STAGE"; return 1
  fi
  if [ ! -e "$STAGE/Payload/$NAME/Frameworks/Flutter.framework/Flutter" ]; then
    echo "!! Payload/$NAME 缺 Flutter.framework/Flutter, 不像 flutter build 的产物" >&2
    rm -rf "$STAGE"; return 1
  fi

  if ! (cd "$STAGE" && zip -qry "$OUT" Payload); then
    echo "!! zip 失败" >&2
    rm -rf "$STAGE"; return 1
  fi
  rm -rf "$STAGE"

  # 第二道（归档层）: 用变量匹配, 不写 `unzip -Z1 | grep -q` ——
  # pipefail 下 grep 提前退出会让 unzip 收 SIGPIPE, 命中也判失败（我在 macOS 上踩过）
  LIST=$(unzip -Z1 "$OUT" || true)
  case $'\n'"$LIST"$'\n' in
    *$'\n'"Payload/$NAME/"*) ;;
    *)
      echo "!! IPA 里 Payload/ 下没有 $NAME 目录, 装机会报 unable to locate app bundle" >&2
      return 1 ;;
  esac
  echo "    结构 ok (Payload/$NAME/${EXE:+, 可执行 $EXE})"
  echo "    $OUT"
}

build_ios() {
  if [[ "$(uname)" != "Darwin" ]]; then
    echo "!! iOS 构建需要 macOS + Xcode, 当前系统 $(uname) 跳过" >&2
    return 0
  fi
  echo "==> 构建 iOS (未签名)"
  (cd client \
    && flutter pub get \
    && flutter gen-l10n \
    && dart run flutter_launcher_icons \
    && flutter build ios --release --no-codesign)

  echo "==> 定位 .app"
  APP=client/build/ios/iphoneos/Runner.app
  [ -d "$APP" ] || APP=$(find client/build/ios -type d -name Runner.app | head -1)
  [ -n "$APP" ] && [ -d "$APP" ] || { echo "!! 没找到 Runner.app, 构建没出产物" >&2; exit 1; }
  echo "    $APP"

  echo "==> 打包 IPA"
  package_ipa "$APP" "$ROOT/dist/hitest-ios-unsigned.ipa"
  ls -la dist/
}

case "$TARGET" in
  go)   build_go ;;
  ios)  build_ios ;;
  icon) node scripts/gen_icon.js ;;
  all)
    build_go
    # iOS 仅在 macOS 上跳过, 其他失败要传出来(坏包不能静默推上去)
    build_ios
    ;;
esac

echo "==> 完成"
