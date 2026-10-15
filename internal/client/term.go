package client

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

// Con: 交互终端 —— 本地终端 raw 模式 ↔ server WS ↔ agent pty。
// 体验与 ssh 相同: ctrl-c 直达远端, 远端 exit 后本地退出。
//
// WS 协议（与浏览器 xterm 同一条路）:
//   首条文本消息 = {"cols":..,"rows":..,"cmd":"/bin/bash -l"}
//   二进制消息  = [1B kind][payload]  kind: 0=数据 1=resize(4B) 2=关闭
func (c *Config) Con(machine string, cmd string) error {
	if cmd == "" {
		cmd = "/bin/bash -l"
	}
	wsURL := httpToWS(c.machineURL(machine)) + "/terminal/ws"
	header := map[string][]string{}
	if c.Token != "" {
		header["X-Token"] = []string{c.Token}
	}
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return fmt.Errorf("连接 %s: %w", wsURL, err)
	}
	defer ws.Close()

	// 本地终端 raw 模式
	fd := int(os.Stdin.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("终端非 tty? %w", err)
	}
	defer term.Restore(fd, oldState)

	// open 消息（含初始窗口）
	cols, rows := winSize(fd)
	open, _ := json.Marshal(map[string]any{"cols": cols, "rows": rows, "cmd": cmd})
	if err := ws.WriteMessage(websocket.TextMessage, open); err != nil {
		return err
	}

	// SIGWINCH → resize 帧
	winch := make(chan os.Signal, 1)
	signal.Notify(winch, syscall.SIGWINCH)
	go func() {
		for range winch {
			c2, r2 := winSize(fd)
			buf := make([]byte, 5)
			buf[0] = 1 // FrameResize
			buf[1], buf[2] = byte(c2>>8), byte(c2)
			buf[3], buf[4] = byte(r2>>8), byte(r2)
			ws.WriteMessage(websocket.BinaryMessage, buf)
		}
	}()

	// WS → stdout
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			mt, data, err := ws.ReadMessage()
			if err != nil {
				return
			}
			if mt == websocket.BinaryMessage && len(data) > 0 && data[0] == 2 { // close
				return
			}
			if mt == websocket.BinaryMessage && len(data) > 0 {
				os.Stdout.Write(data[1:])
			}
		}
	}()

	// stdin → WS（raw 模式下按字节直达, 与 ssh 相同）
	buf := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			break
		}
		frame := make([]byte, 1+n)
		frame[0] = 0
		copy(frame[1:], buf[:n])
		if err := ws.WriteMessage(websocket.BinaryMessage, frame); err != nil {
			break
		}
	}
	ws.Close()
	<-done
	return nil
}

func winSize(fd int) (cols, rows uint16) {
	var sz struct {
		Row, Col, X, Y uint16
	}
	_, _, _ = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&sz)))
	return sz.Col, sz.Row
}

func httpToWS(u string) string {
	switch {
	case hasPrefix(u, "https://"):
		return "wss://" + u[len("https://"):]
	case hasPrefix(u, "http://"):
		return "ws://" + u[len("http://"):]
	}
	return u
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }
