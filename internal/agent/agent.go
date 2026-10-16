// Package agent: agent 角色。
//
// 职责:
//  1. 本机起 gin + 功能模块, 监听 unix socket（功能代码零传输耦合）
//  2. 主动外连 server, yamux 会话上按 stream 类型分发:
//     http → 桥接到本机 unix socket; term → modules/terminal.ServeStream
//  3. 断线指数退避重连
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"hitest/internal/debugx"
	"hitest/internal/modules"
	"hitest/internal/modules/terminal"
	"hitest/internal/store"
	"hitest/internal/transport"
)

type Options struct {
	ID       string // 机器 ID（默认 hostname）
	Token    string
	Server   string // ws://host:port 或 wss://
	SockPath string // unix socket 路径（空则自动生成）
	DataDir  string // sqlite 存放目录（模块持久化用）
	Debug    bool   // 内存/Goroutine 快照 + 回环 pprof（--debug）
}

// LocalStart 只启动本机 API（gin + 模块, 监听 unix socket）。
// server 进程的内置本机 agent 用它; 独立 agent 用 Start。
func LocalStart(ctx context.Context, opts Options) (string, error) {
	if opts.SockPath == "" {
		opts.SockPath = filepath.Join(os.TempDir(), fmt.Sprintf("hitest-agent-%d.sock", os.Getpid()))
	}
	_ = os.Remove(opts.SockPath)

	// 模块持久化: 独立 agent 用 DataDir/agent.db; 内置本机 agent 复用 server 的 store
	var st *store.Store
	if opts.DataDir != "" {
		s, err := store.Open(filepath.Join(opts.DataDir, "agent.db"))
		if err != nil {
			return "", fmt.Errorf("打开 agent 数据库: %w", err)
		}
		st = s
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	api := r.Group("/api/v1")
	for _, m := range modules.All() {
		m.RegisterRoutes(api, st)
	}
	ln, err := net.Listen("unix", opts.SockPath)
	if err != nil {
		return "", fmt.Errorf("监听 unix socket: %w", err)
	}
	srv := &http.Server{Handler: r}
	go func() {
		<-ctx.Done()
		srv.Close()
	}()
	go func() {
		if err := srv.Serve(ln); err != nil && ctx.Err() == nil {
			log.Printf("[agent] http server 退出: %v", err)
		}
	}()
	log.Printf("[agent] 本机 API 就绪: unix://%s", opts.SockPath)
	return opts.SockPath, nil
}

// Start 阻塞运行独立 agent（本机 API + 外连循环）直到 ctx 取消。
func Start(ctx context.Context, opts Options) error {
	if opts.ID == "" {
		h, _ := os.Hostname()
		opts.ID = h
	}
	// 独立 agent 进程才开 debug；server 进程里的内置本机 agent 由 server 那侧统一报（同进程, 一份快照就够）
	if opts.Debug {
		debugx.Start(ctx, debugx.Options{
			Role: "agent",
			Addr: "127.0.0.1:6062",
			Extra: func() string {
				return "id=" + opts.ID
			},
		})
	}
	sock, err := LocalStart(ctx, opts)
	if err != nil {
		return err
	}
	opts.SockPath = sock

	// 外连循环
	backoff := time.Second
	for ctx.Err() == nil {
		start := time.Now()
		err := serveConn(ctx, opts)
		if ctx.Err() != nil {
			return nil
		}
		if err != nil {
			log.Printf("[agent] 连接断开: %v", err)
		}
		if time.Since(start) > time.Minute {
			backoff = time.Second // 长连接成功过, 重置退避
		}
		jitter := time.Duration(rand.Int63n(int64(backoff / 2)))
		log.Printf("[agent] %v 后重连", backoff+jitter)
		select {
		case <-time.After(backoff + jitter):
		case <-ctx.Done():
			return nil
		}
		if backoff < 30*time.Second {
			backoff *= 2
		}
	}
	return nil
}

// serveConn: 连接 → 注册 → yamux 分发循环, 直到出错
func serveConn(ctx context.Context, opts Options) error {
	ws, _, err := websocket.DefaultDialer.Dial(opts.Server+"/agent/ws", nil)
	if err != nil {
		return fmt.Errorf("拨号 %s: %w", opts.Server, err)
	}
	defer ws.Close()

	hostname, _ := os.Hostname()
	reg := transport.Register{
		ID:       opts.ID,
		Token:    opts.Token,
		Hostname: hostname,
		OS:       runtime.GOOS + "/" + runtime.GOARCH,
	}
	ws.WriteJSON(reg)

	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, data, err := ws.ReadMessage()
	if err != nil {
		return fmt.Errorf("读注册应答: %w", err)
	}
	var ack transport.RegisterAck
	if json.Unmarshal(data, &ack) != nil || !ack.OK {
		return fmt.Errorf("注册被拒绝: %s", ack.Error)
	}
	_ = ws.SetReadDeadline(time.Time{})
	log.Printf("[agent] 已连接 %s (id=%s)", opts.Server, opts.ID)

	sess, err := transport.NewClientSession(transport.NewWSNetConn(ws))
	if err != nil {
		return fmt.Errorf("yamux 会话: %w", err)
	}
	defer sess.Close()

	// sess.Accept 是纯阻塞调用, 不看 ctx。没有这个 watcher 的话, systemd 发的
	// SIGTERM 只取消 ctx 却唤不醒 Accept —— 进程要等 TimeoutStopSec 到点被 SIGKILL。
	// 关会话让 Accept 立即返回错误, 外层 serveConn 带 ctx.Err() 检查正常退出。
	go func() {
		<-ctx.Done()
		sess.Close()
	}()

	for {
		typ, conn, err := sess.Accept()
		if err != nil {
			return err
		}
		switch typ {
		case transport.StreamHTTP:
			go bridgeHTTP(conn, opts.SockPath)
		case transport.StreamTerm:
			go func() { _ = terminal.ServeStream(conn) }()
		default:
			conn.Close() // 未知类型, 直接关闭
		}
	}
}

// bridgeHTTP: stream[http] ↔ 本机 unix socket 双向字节泵
func bridgeHTTP(c net.Conn, sockPath string) {
	defer c.Close()
	up, err := net.Dial("unix", sockPath)
	if err != nil {
		log.Printf("[agent] 桥接失败: %v", err)
		return
	}
	defer up.Close()
	done := make(chan struct{}, 2)
	go func() { io.Copy(up, c); done <- struct{}{} }()
	go func() { io.Copy(c, up); done <- struct{}{} }()
	<-done
	// 一个方向结束（另一端 EOF / 客户端断开）后, 另一个方向的 Copy 还会堵在那里。
	// 两边都关掉让它带着错误退出, 否则每断一次请求就永久滞留一个 io.Copy goroutine。
	c.Close()
	up.Close()
}
