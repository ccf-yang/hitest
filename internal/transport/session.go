package transport

import (
	"fmt"
	"io"
	"net"

	"github.com/hashicorp/yamux"
)

// StreamType 是 stream 的第一个字节
type StreamType byte

const (
	StreamHTTP StreamType = 1 // 原始 HTTP/1.1 字节泵
	StreamTerm StreamType = 2 // 终端流（帧格式见 modules/terminal）
)

// Session 是一条 agent 连接上的 yamux 会话。
// 双端都可开 stream；stream 首字节 = StreamType。
type Session struct {
	ys *yamux.Session
}

func muxConfig() *yamux.Config {
	conf := yamux.DefaultConfig()
	conf.KeepAliveInterval = 20 * 1e9 // 20s, yamux 内置保活, 死链由它发现
	conf.LogOutput = io.Discard
	return conf
}

// NewServerSession: server 侧（接收 agent 外连的一方）
func NewServerSession(c net.Conn) (*Session, error) {
	ys, err := yamux.Server(c, muxConfig())
	if err != nil {
		return nil, err
	}
	return &Session{ys: ys}, nil
}

// NewClientSession: agent 侧（主动拨号的一方）
func NewClientSession(c net.Conn) (*Session, error) {
	ys, err := yamux.Client(c, muxConfig())
	if err != nil {
		return nil, err
	}
	return &Session{ys: ys}, nil
}

// Open 打开一条指定类型的 stream（写类型字节）
func (s *Session) Open(t StreamType) (net.Conn, error) {
	c, err := s.ys.Open()
	if err != nil {
		return nil, err
	}
	if _, err := c.Write([]byte{byte(t)}); err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

// Accept 接收一条 stream（读掉类型字节）
func (s *Session) Accept() (StreamType, net.Conn, error) {
	c, err := s.ys.Accept()
	if err != nil {
		return 0, nil, err
	}
	var b [1]byte
	if _, err := io.ReadFull(c, b[:]); err != nil {
		c.Close()
		return 0, nil, fmt.Errorf("读 stream 类型: %w", err)
	}
	return StreamType(b[0]), c, nil
}

func (s *Session) Close() error { return s.ys.Close() }
func (s *Session) IsClosed() bool { return s.ys.IsClosed() }
