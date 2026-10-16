// Package debugx: --debug 打开时的自检输出，只观测、不改变任何业务行为。
//
// 开三项：
//  1. 每 60s 往 stdout 写一行内存/资源快照（RSS、堆、GC、Goroutine、fd、pts）
//     —— journald 落盘，事后能按时间轴对齐「突增发生在什么时候」。
//  2. 在 127.0.0.1 上挂 net/http/pprof（默认 server :6061 / agent :6062）
//     —— 只绑回环，局域网摸不到；随时抓堆、Goroutine、block 剖面。
//  3. 打开 block/mutex 剖面采样（1ms / 1%），用来抓「卡在锁或 IO 上」这类僵死。
//
// 不开 --debug 时本包一行都不会跑，pprof 也不监听。
package debugx

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof" // 注册进 DefaultServeMux, 只由本包的回环监听对外可见
	"os"
	"runtime"
	"strings"
	"time"
)

type Options struct {
	Role  string        // 日志前缀: server / agent
	Addr  string        // pprof 地址, 空则用 DefaultAddr
	Every time.Duration // 快照间隔, 0 = 60s
	Extra func() string // 附加字段（在线机器数、终端会话数等）；可为 nil
}

// Start 起快照循环与 pprof 监听，随 ctx 结束而停。pprof 端口被占只记日志，不影响主服务。
func Start(ctx context.Context, o Options) {
	if o.Role == "" {
		o.Role = "server"
	}
	if o.Addr == "" {
		o.Addr = "127.0.0.1:6061"
	}
	if !strings.HasPrefix(o.Addr, "127.0.0.1:") && !strings.HasPrefix(o.Addr, "[::1]:") {
		// 硬拦一道：pprof 无鉴权，绝不允许监听非回环地址
		log.Printf("[debug][%s] 拒绝在非回环地址 %s 开 pprof，已跳过监听", o.Role, o.Addr)
		o.Addr = ""
	}
	if o.Every <= 0 {
		o.Every = 60 * time.Second
	}

	if o.Addr != "" {
		go func() {
			srv := &http.Server{Addr: o.Addr, ReadHeaderTimeout: 5 * time.Second}
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[debug][%s] pprof 监听 %s 失败: %v", o.Role, o.Addr, err)
			}
		}()
		// 采样率只在 debug 下开，常态零开销
		runtime.SetBlockProfileRate(int(time.Millisecond))
		runtime.SetMutexProfileFraction(100)
		log.Printf("[debug][%s] pprof: http://%s/debug/pprof/  快照间隔 %s", o.Role, o.Addr, o.Every)
	}

	go loop(ctx, o)
}

func loop(ctx context.Context, o Options) {
	var lastGC uint32
	snap(o, 0) // 启动即来一条，方便和部署时间对齐
	t := time.NewTicker(o.Every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			d := m.NumGC - lastGC
			lastGC = m.NumGC
			snap(o, d)
		}
	}
}

func snap(o Options, gcDelta uint32) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fds, pts := countFDs()

	extra := ""
	if o.Extra != nil {
		if s := strings.TrimSpace(o.Extra()); s != "" {
			extra = " " + s
		}
	}

	log.Printf("[debug][%s] rss=%s heap=%s inuse=%s sys=%s stack=%s gor=%d fd=%d pts=%d gc+%d pauseMax=%s%s",
		o.Role,
		mb(rssKB()), mb(m.HeapAlloc/1024), mb(m.HeapInuse/1024), mb(m.Sys/1024), mb(m.StackInuse/1024),
		runtime.NumGoroutine(), fds, pts, gcDelta, gcPauseMax(&m), extra)
}

// mb: KB → "12.3MB"
func mb(kb uint64) string { return fmt.Sprintf("%.1fMB", float64(kb)/1024) }

// gcPauseMax: 最近一轮 GC 的最大 pause
func gcPauseMax(m *runtime.MemStats) string {
	if m.NumGC == 0 {
		return "0ms"
	}
	var max uint64
	for i := 0; i < 8; i++ {
		if v := m.PauseNs[(m.NumGC+255-uint32(i))%256]; v > max {
			max = v
		}
	}
	return fmt.Sprintf("%.2fms", float64(max)/1e6)
}

// rssKB: /proc/self/status 的 VmRSS（真实驻留, 比 MemStats.Sys 更贴近“占了多少内存”）。
// 非 Linux 取不到就返回 0, 此时日志里显示 0.0MB, 不影响其它字段。
func rssKB() uint64 {
	b, err := os.ReadFile("/proc/self/status")
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(b), "\n") {
		rest, ok := strings.CutPrefix(line, "VmRSS:")
		if !ok {
			continue
		}
		var kb uint64
		for _, ch := range strings.TrimSpace(rest) {
			if ch < '0' || ch > '9' {
				break
			}
			kb = kb*10 + uint64(ch-'0')
		}
		return kb
	}
	return 0
}

func countFDs() (fds, pts int) {
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return 0, 0
	}
	for _, e := range entries {
		fds++
		if link, err := os.Readlink("/proc/self/fd/" + e.Name()); err == nil &&
			strings.HasPrefix(link, "/dev/pts") {
			pts++
		}
	}
	return
}
