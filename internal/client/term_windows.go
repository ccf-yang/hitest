//go:build windows

package client

import (
	"time"

	"golang.org/x/term"
)

const (
	defaultCols = 80
	defaultRows = 24
)

// winSize: Windows 控制台没有 TIOCGWINSZ, 用屏幕缓冲区信息。
func winSize(fd int) (cols, rows uint16) {
	w, h, err := term.GetSize(fd)
	if err != nil || w <= 0 || h <= 0 {
		return defaultCols, defaultRows
	}
	return uint16(w), uint16(h)
}

// watchResize: Windows 没有窗口尺寸信号（Ctrl+C 事件也不带尺寸）,
// 只能轮询控制台尺寸; 变化时才回调。
func watchResize(fd int, send func(cols, rows uint16)) func() {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		lastCols, lastRows := winSize(fd)
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				cols, rows := winSize(fd)
				if cols != lastCols || rows != lastRows {
					lastCols, lastRows = cols, rows
					send(cols, rows)
				}
			}
		}
	}()
	return func() { close(done) }
}
