# hitest

家庭/边缘 Linux 机器管理平台。一个二进制，三种角色：

```
浏览器 / CLI / App
      │ HTTP + 终端WS
      ▼
hitest serve（常开机器, 内置本机 agent）
  ├─ /api/v1/local/*    → 本机
  ├─ /api/v1/m/{id}/*   → 隧道反代到远程 agent
  └─ /agent/ws          ← agent 主动外连（WS + yamux, 穿 NAT）

hitest agent -s ws://hub -t <token>   （NAS / Armbian 上）
```

架构要点见 `DESIGN.md`：**功能 = HTTP handler**（唯一例外是终端 stream 型），
加功能 = `internal/modules/` 加目录 + 注册表加一行。

## 构建

依赖: Go 1.26+, Node.js 18+（iOS 需 macOS + Xcode + Flutter）。

```bash
make build              # = build.sh all → bin/hitest-linux-{amd64,arm64} (macOS 上附 IPA)
make build_go           # 只构建 Go 二进制（自动嵌前端）
make test               # 全量测试（含端到端）
make e2e                # 只跑端到端
scripts/build.sh go --arch arm64   # 只构建指定架构
scripts/build.sh ios    # 未签名 iOS IPA（仅 macOS）
```

CI: GitHub Actions（`.github/workflows/build.yml`）—— Actions 页手动触发或推 `v*` tag，
自动构建 linux 二进制 + iOS IPA；tag 触发时产物附到 Release。

## 使用

```bash
# 1. 中心（首次启动自动生成 config.yaml, 含注册 token）
./hitest serve

# 2. 节点（拷二进制过去）
sudo sh scripts/install.sh -s ws://<server-ip>:8080 -t <token> -n nas1

# 3. CLI
./hitest login http://<server-ip>:8080
./hitest ls nas1 /opt
./hitest exec nas1 "systemctl status nginx" --json   # AI agent 工具形态
./hitest cp app.tar.gz nas1:/opt/app/                # 多机: nas1,nas2:/path
./hitest con nas1                                     # 交互终端（像 ssh）
```

## 配置 (config.yaml)

```yaml
port: 8080
token: "<agent 注册用, 首次启动自动生成>"
access_token: ""     # 非空则 Web/API 需带 X-Token 头
cert_file: ""        # TLS（可选）
key_file: ""
data_dir: "data"
```

## Flutter 客户端 (client/)

```bash
cd client
flutter pub get && flutter gen-l10n
dart run flutter_launcher_icons   # 首次构建前生成 AppIcon
flutter build ios --release --no-codesign
```

- 服务器地址填 `http://<server-ip>:8080`；apiKey 任意（当前服务端不校验，仅本地保存）
- 客户端带 `X-Hitest-Envelope: 1` 头时服务端返回 legacy `{code,message,data}` 信封（`internal/server/envelope.go`），Web 前端不受影响
- 白屏修复记录：旧 AppDelegate.swift 内嵌 SwiftUI `@main RunnerApp` + 原生 TabView（对应已删功能），其启动画面等待已不存在的 `getUIRenderMode` channel 回调，导致永远白屏。已恢复标准 FlutterAppDelegate 生命周期
