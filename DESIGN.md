# hitest 统一架构设计 v1

> 目标：一个简单、易维护、可扩展的家庭/边缘 Linux 机器管理平台。
> 覆盖 homeops（多机+终端+文件）、hitest（监控+工具箱+Flutter App）、syncd（文件同步）的合并蓝图。
> 日期：2026-10-14

## 0. 经验教训（设计依据）

| 来源 | 教训 |
|---|---|
| homeops | 自研 `[connID][payload]` 协议是包袱：会话路由表、fs 专用分块、resize 带外消息，每个功能都碰传输层，还出过早到输入竞态 |
| hitest | "功能 = 普通 HTTP handler"极好（gin 层零传输耦合）；但 unix socket 只能本机；cmux 三模式/xpack/空转 DB 是负担 |
| syncd | 控制面小 JSON + 数据面标准 HTTP（Range 下载），全程 curl 可调试，简单可靠 |

## 1. 设计原则

1. **功能永远不碰传输层**：任何功能 = HTTP handler；只有真·字节流（终端）用流通道。
2. 一个二进制，三种角色：`hitest serve` / `hitest agent -s ws://hub` /（未来可加）。
3. 一切状态进一个 sqlite；一切配置进一个 yaml。
4. 依赖面小：gin + yamux + gopsutil + sqlite。

## 2. 拓扑

```
浏览器 / App
    │ HTTP + 终端WS
    ▼
hitest serve（常开机器）
  ├─ /api/v1/local/*    → 内置 agent（内存 listener 直连）
  ├─ /api/v1/m/{id}/*   → 隧道反代 → 远程 agent
  ├─ /                  → 内嵌 Vue SPA
  └─ hub: agent 注册表 + sqlite

远程机器: hitest agent -s ws://hub -t <token>
  1 条 WS 长连接（主动外连，穿 NAT），yamux 多路复用：
    stream[type=http] ← 原始 HTTP/1.1 字节泵（API/上传下载/WS 升级全兼容）
    stream[type=term] ← 终端帧
```

## 3. 支柱

### ① 隧道 = HTTP over stream
`httputil.ReverseProxy` + 自定义 `DialContext`（向目标 agent 开 yamux stream，双向泵原始字节）。
- 普通 API：原样工作
- 文件上传/下载：流式 HTTP（multipart / chunked），curl 可测，天然大文件
- WebSocket：Upgrade 后就是双向字节管道，原始字节泵天然承载，零特判
- agent 本机 gin 挂内存 listener，功能代码不知道本地/远程区别

### ② 唯一流型模块 = 终端
帧格式：`[1B kind][payload]`，kind: 0=数据 1=resize(4B cols,rows) 2=对端关闭。
除此之外无任何自定义协议。

### ③ 模块注册表 = 扩展点
```go
type Module interface {
    Name() string
    RegisterRoutes(r gin.IRouter, s *store.Store)
}
type StreamModule interface {   // 终端/未来日志 tail
    Module
    ServeStream(s transport.Stream, header []byte) error
}
```
加功能 = modules/ 加目录 + 注册表加一行。
未来：容器管理(HTTP)、日志 tail(Stream)、**syncd 文件同步(agent 模块)**、通知、AI 代理。

## 4. 目录结构

```
hitest/
├── cmd/hitest/main.go        # cobra: serve / agent / version
├── internal/
│   ├── transport/            # ★ 通信基建（~600行，写完冻结）
│   │   ├── mux.go            #   yamux 封装
│   │   ├── dial.go           #   agent 外连/退避重连/心跳
│   │   ├── hub.go            #   server 侧 agent 注册表（token 校验）
│   │   └── httptunnel.go     #   stream ↔ net.Conn 桥 + ReverseProxy Dialer
│   ├── modules/
│   │   ├── module.go         #   接口 + 注册表
│   │   ├── monitor/          #   ← hitest
│   │   ├── disks/            #   ← hitest
│   │   ├── toolbox/          #   ← hitest
│   │   ├── process/          #   ← hitest
│   │   ├── commands/         #   ← hitest
│   │   ├── files/            #   ← homeops 思路 → 纯 HTTP 流式
│   │   └── terminal/         #   ← homeops（唯一 StreamModule）
│   ├── server/               #   serve 组装：gin、SPA、配置
│   ├── agent/                #   agent 组装：本机 gin + dial
│   ├── client/               #   CLI 客户端：con/exec/cp/ls（同一二进制第三角色）
│   └── store/                #   sqlite 单库（machines/kv/metrics）
├── web/                      #   Vue3 SPA（homeops 壳 + 监控页）
├── client/                   #   Flutter iOS（后期，同一套 API）
├── scripts/install.sh
└── Makefile
```

## 5. CLI 客户端（同一二进制的第三种角色）

`hitest con/exec/cp/ls` —— 与 serve/agent 同一二进制，客户端只是 HTTP + 一条终端 WS 的消费者：

- `hitest login http://hub:8080 -t <token>` → 存 `~/.hitest.yaml`（server 地址 + token + 机器别名）
- `hitest con nas1` → 本地终端 raw 模式 + WS 直通 `/api/v1/m/nas1/terminal/ws` + SIGWINCH resize（同浏览器 xterm 一条路）
- `hitest exec nas1 "cmd" --json` → 非交互执行，结构化输出 + exit code（**AI agent 的工具形态**）
- `hitest cp file nas1:/path` / `cp -r dir nas1,nas2:/path` → 流式 HTTP 上传，多机 goroutine 并行

代码位置：`internal/client/`（config.go / term.go / files.go / exec.go）。
未来 AI 接管 = 授予 CLI 的 exec/ls/cp（确定性工具），con 仅人工接管用。

## 6. 其余决策

- **API**：`/api/v1/local/*` 与 `/api/v1/m/{machine}/*`；`local` = 内置 agent 机器 ID，前端无感。
- **存储**：单库起步 3 表（machines / kv / metric_samples），模块登记各自 AutoMigrate。
- **鉴权**：agent↔server token；Web/App 默认无鉴权（局域网），留 `access_token` 配置项一行开启。
- **配置**：server 一个 yaml（listen/token/access_token/modules 启停）；agent 仅 `-s -t`。
- **废弃**：homeops 自研协议/会话路由/hub；hitest 的 cmux 三模式、xpack、unix-socket-only、disableAuth、5 个空转 DB、ApiPaths 死代码。
- **复用资产**：hitest 模块业务逻辑（纯 Go）；homeops pty 终端 + Vue 壳 + e2e 测试方法学。

## 7. 可维护性自检

- 简单：传输（冻结）+ 模块（平铺）两层心智模型
- 好维护：全程 curl 可调试；分层定位；每模块 handler 测试 + 过真实隧道的 e2e
- 可扩展：HTTP 功能 = 加目录；流功能 = StreamModule；独立工具可模块化并入
