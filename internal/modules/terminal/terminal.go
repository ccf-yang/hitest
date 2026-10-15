// Package terminal: 交互终端 —— 全系统唯一的非 HTTP（stream 型）功能。
//
// stream[term] 布局:
//
//	头部: [2B cmdLen][cmd][2B cols][2B rows]
//	帧:   [1B kind][2B len BE][payload]
//	      kind 0 = 数据(pty 输入/输出)
//	      kind 1 = resize(payload 4B: cols u16, rows u16)
//	      kind 2 = 关闭
//
// WS 侧（浏览器/CLI）不带头部和长度前缀（消息边界即帧界）：
//
//	WS 消息 = [1B kind][payload]，open 参数为首条 JSON 文本消息。
package terminal

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/creack/pty"

	"hitest/internal/transport"
)

// ---------- PTY 原语（server 本机模式与 agent 远程模式共用） ----------

// PTY 是一个终端会话。Start 同步返回（早到输入进缓冲, 不丢失）。
type PTY struct {
	mu        sync.Mutex
	f         *os.File
	cmd       *exec.Cmd
	input     chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

func Start(cols, rows uint16, cmdLine string) (*PTY, error) {
	parts := strings.Fields(cmdLine)
	if len(parts) == 0 {
		parts = []string{"/bin/bash", "-l"}
	}
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	f, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}
	_ = pty.Setsize(f, &pty.Winsize{Rows: rows, Cols: cols})

	p := &PTY{
		f:     f,
		cmd:   cmd,
		input: make(chan []byte, 64),
		done:  make(chan struct{}),
	}
	go p.pumpInput()
	return p, nil
}

func (p *PTY) pumpInput() {
	for {
		select {
		case b := <-p.input:
			if _, err := p.f.Write(b); err != nil {
				return
			}
		case <-p.done:
			return
		}
	}
}

// Read 读 pty 输出（阻塞, 供桥接循环调用）
func (p *PTY) Read(b []byte) (int, error) { return p.f.Read(b) }

func (p *PTY) Write(b []byte) (int, error) {
	select {
	case p.input <- b:
		return len(b), nil
	case <-p.done:
		return 0, io.ErrClosedPipe
	}
}

func (p *PTY) Resize(cols, rows uint16) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.f == nil {
		return nil
	}
	return pty.Setsize(p.f, &pty.Winsize{Rows: rows, Cols: cols})
}

func (p *PTY) Close() {
	p.closeOnce.Do(func() {
		close(p.done)
		if p.cmd.Process != nil {
			_ = p.cmd.Process.Kill()
		}
		_ = p.cmd.Wait()
		_ = p.f.Close()
	})
}

// ---------- 流帧编解码 ----------

const (
	FrameData   byte = 0
	FrameResize byte = 1
	FrameClose  byte = 2
)

func WriteFrame(w io.Writer, kind byte, payload []byte) error {
	head := make([]byte, 3, 3+len(payload))
	head[0] = kind
	binary.BigEndian.PutUint16(head[1:], uint16(len(payload)))
	_, err := w.Write(append(head, payload...))
	return err
}

func ReadFrame(r io.Reader) (kind byte, payload []byte, err error) {
	head := make([]byte, 3)
	if _, err = io.ReadFull(r, head); err != nil {
		return
	}
	kind = head[0]
	n := binary.BigEndian.Uint16(head[1:])
	payload = make([]byte, n)
	_, err = io.ReadFull(r, payload)
	return
}

// writeHeader: 头部 = [2B cmdLen][cmd][2B cols][2B rows]
func writeHeader(w io.Writer, cmd string, cols, rows uint16) error {
	if len(cmd) > 4096 {
		return fmt.Errorf("cmd too long")
	}
	buf := make([]byte, 2+len(cmd)+4)
	binary.BigEndian.PutUint16(buf, uint16(len(cmd)))
	copy(buf[2:], cmd)
	binary.BigEndian.PutUint16(buf[2+len(cmd):], cols)
	binary.BigEndian.PutUint16(buf[2+len(cmd)+2:], rows)
	_, err := w.Write(buf)
	return err
}

// ---------- server 侧: 打开远程终端流 ----------

// OpenRemote 在指定 agent 会话上打开一条终端流（server 进程里调用）。
// 返回的 conn 之后按帧协议读写。
func OpenRemote(sess *transport.Session, cmd string, cols, rows uint16) (net.Conn, error) {
	c, err := sess.Open(transport.StreamTerm)
	if err != nil {
		return nil, err
	}
	if err := writeHeader(c, cmd, cols, rows); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// ---------- agent 侧: stream[term] 处理器 ----------

// ServeStream 处理一条来自 server 的终端流（在 agent 进程里运行）
func ServeStream(c net.Conn) error {
	defer c.Close()

	// 头部: [2B cmdLen][cmd][2B cols][2B rows]
	var head [2]byte
	if _, err := io.ReadFull(c, head[:]); err != nil {
		return err
	}
	cmdLen := binary.BigEndian.Uint16(head[:])
	cmdBytes := make([]byte, cmdLen)
	if _, err := io.ReadFull(c, cmdBytes); err != nil {
		return err
	}
	var win [4]byte
	if _, err := io.ReadFull(c, win[:]); err != nil {
		return err
	}
	cols := binary.BigEndian.Uint16(win[0:])
	rows := binary.BigEndian.Uint16(win[2:])

	p, err := Start(cols, rows, string(cmdBytes))
	if err != nil {
		_ = WriteFrame(c, FrameClose, []byte(err.Error()))
		return nil
	}
	defer p.Close()

	// pty 输出 → 数据帧
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := p.Read(buf)
			if n > 0 {
				if werr := WriteFrame(c, FrameData, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				_ = WriteFrame(c, FrameClose, nil)
				return
			}
		}
	}()

	// 帧 → pty 输入/resize
	for {
		kind, payload, err := ReadFrame(c)
		if err != nil {
			return nil
		}
		switch kind {
		case FrameData:
			if _, err := p.Write(payload); err != nil {
				return nil
			}
		case FrameResize:
			if len(payload) >= 4 {
				_ = p.Resize(binary.BigEndian.Uint16(payload),
					binary.BigEndian.Uint16(payload[2:]))
			}
		case FrameClose:
			return nil
		}
	}
}
