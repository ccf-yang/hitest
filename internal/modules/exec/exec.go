// Package exec: 非交互命令执行 —— CLI `hitest exec` 与 AI agent 的工具形态。
// POST /exec/run {"cmd":"...", "timeout_sec":10}
//   → {"code":0, "stdout":"...", "stderr":"...", "dur_ms":123}
package exec

import (
	"bytes"
	"context"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hitest/internal/store"
)

const maxOutput = 1 << 20 // 单路输出上限 1MB, 防止失控命令撑爆内存

type ExecModule struct{}

func (m *ExecModule) Name() string { return "exec" }

func (m *ExecModule) RegisterRoutes(api gin.IRouter, _ *store.Store) {
	api.POST("/exec/run", m.run)
}

type RunResult struct {
	Code   int    `json:"code"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	DurMs  int64  `json:"dur_ms"`
}

func (m *ExecModule) run(c *gin.Context) {
	var req struct {
		Cmd        string `json:"cmd"`
		TimeoutSec int    `json:"timeout_sec"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Cmd == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing cmd"})
		return
	}
	timeout := time.Duration(req.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "bash", "-c", req.Cmd)
	// 限量捕获: 防止失控命令(yes / cat 大文件)在超时前撑爆内存。
	// 写满后丢弃后续输出, 管道反压会让进程阻塞直至被超时杀掉。
	stdout := &limitedWriter{limit: maxOutput}
	stderr := &limitedWriter{limit: maxOutput}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	err := cmd.Run()
	res := RunResult{DurMs: time.Since(start).Milliseconds()}
	res.Code = 0
	res.Stdout = stdout.String()
	res.Stderr = stderr.String()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.Code = exitErr.ExitCode()
		} else {
			// 启动失败/超时: code = -1, 错误放 stderr
			res.Code = -1
			res.Stderr = truncate(stderr.String() + err.Error())
		}
	}
	c.JSON(http.StatusOK, res)
}

func truncate(s string) string {
	if len(s) > maxOutput {
		return s[:maxOutput] + "\n...[truncated]"
	}
	return s
}

// limitedWriter 捕获至多 limit 字节, 之后丢弃(计入 discarded)
type limitedWriter struct {
	buf       bytes.Buffer
	limit     int
	discarded int64
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if room := w.limit - w.buf.Len(); room > 0 {
		if len(p) <= room {
			return w.buf.Write(p)
		}
		w.buf.Write(p[:room])
		w.discarded += int64(len(p) - room)
		return len(p), nil // 谎报全写, 避免命令因 EPIPE 提前退出
	}
	w.discarded += int64(len(p))
	return len(p), nil
}

func (w *limitedWriter) String() string {
	s := w.buf.String()
	if w.discarded > 0 {
		return s + "\n...[truncated " + strconv.FormatInt(w.discarded, 10) + " bytes]"
	}
	return s
}
