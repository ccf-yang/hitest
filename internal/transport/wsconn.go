// Package transport 是全系统唯一的通信基建。
//
// 模型：agent 主动外连一条 WebSocket，其上跑 yamux 多路复用；
// 每条 stream 的第一个字节是类型(StreamType)，之后才是载荷：
//
//	stream[http]  = 原始 HTTP/1.1 字节双向泵（API/上传下载/WS升级全兼容）
//	stream[term]  = 终端流（头 + 帧，见 modules/terminal）
//
// 本文件: 把 gorilla WebSocket 适配成 net.Conn，供 yamux 使用。
package transport

import (
	"io"
	"net"
	"time"

	"github.com/gorilla/websocket"
)

// WSNetConn 把 WebSocket 连接包装为 net.Conn（二进制消息 = 字节流）。
type WSNetConn struct {
	c *websocket.Conn
	r io.Reader
}

func NewWSNetConn(c *websocket.Conn) *WSNetConn {
	return &WSNetConn{c: c}
}

func (w *WSNetConn) Read(b []byte) (int, error) {
	for {
		if w.r == nil {
			mt, r, err := w.c.NextReader()
			if err != nil {
				return 0, err
			}
			if mt != websocket.BinaryMessage {
				continue
			}
			w.r = r
		}
		n, err := w.r.Read(b)
		if err == io.EOF {
			w.r = nil
			if n == 0 {
				continue
			}
			err = nil
		}
		return n, err
	}
}

func (w *WSNetConn) Write(b []byte) (int, error) {
	if err := w.c.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (w *WSNetConn) Close() error         { return w.c.Close() }
func (w *WSNetConn) LocalAddr() net.Addr  { return w.c.LocalAddr() }
func (w *WSNetConn) RemoteAddr() net.Addr { return w.c.RemoteAddr() }

func (w *WSNetConn) SetDeadline(t time.Time) error {
	if err := w.c.SetReadDeadline(t); err != nil {
		return err
	}
	return w.c.SetWriteDeadline(t)
}

func (w *WSNetConn) SetReadDeadline(t time.Time) error  { return w.c.SetReadDeadline(t) }
func (w *WSNetConn) SetWriteDeadline(t time.Time) error { return w.c.SetWriteDeadline(t) }
