//go:build !windows

package client

import (
	"os"
	"os/signal"
	"syscall"
	"unsafe"
)

// winSize: TIOCGWINSZ 取当前窗口尺寸。
func winSize(fd int) (cols, rows uint16) {
	var sz struct {
		Row, Col, X, Y uint16
	}
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return sz.Col, sz.Row
}

// watchResize 订阅 SIGWINCH, 尺寸变化时回调; 返回的函数停止订阅。
func watchResize(fd int, send func(cols, rows uint16)) func() {
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case <-winch:
				cols, rows := winSize(fd)
				send(cols, rows)
			}
		}
	}()
	return func() {
		signal.Stop(winch)
		close(done)
	}
}
