package commands

import (
	"bytes"
	"io"
)

// inFlightOutputLimit 是子进程输出在内存里的驻留上限。
// quickCommandOutputLimit 只在返回前截断字符串, 挡不住 `yes`、`cat 大文件`
// 这类命令在 120s 超时前把几十 GB 写进缓冲区 —— 盒子只有 3.7GB, 一次就 OOM。
// 这里让缓冲区到顶即止, 之后只计数不存字节（丢弃而非反压, 因为白名单命令是
// 人点的, 阻塞到超时反而拿不到结果）。
const inFlightOutputLimit = 4 << 20

type cappedWriter struct {
	buf       bytes.Buffer
	limit     int
	discarded int64
}

var _ io.Writer = (*cappedWriter)(nil)

func (w *cappedWriter) Write(p []byte) (int, error) {
	if room := w.limit - w.buf.Len(); room > 0 {
		if len(p) <= room {
			return w.buf.Write(p)
		}
		w.buf.Write(p[:room])
		w.discarded += int64(len(p) - room)
		return len(p), nil
	}
	w.discarded += int64(len(p))
	return len(p), nil // 谎报全写, 避免子进程因 EPIPE 提前退出
}

func (w *cappedWriter) String() string { return w.buf.String() }

func (w *cappedWriter) Truncated() int64 { return w.discarded }
