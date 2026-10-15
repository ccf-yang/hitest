package terminal

import (
	"bytes"
	"io"
	"net"
	"testing"
)

func TestFrameRoundtrip(t *testing.T) {
	var buf bytes.Buffer
	payload := []byte("terminal output bytes")
	if err := WriteFrame(&buf, FrameData, payload); err != nil {
		t.Fatal(err)
	}
	if err := WriteFrame(&buf, FrameResize, []byte{0, 80, 0, 24}); err != nil {
		t.Fatal(err)
	}

	kind, got, err := ReadFrame(&buf)
	if err != nil || kind != FrameData || !bytes.Equal(got, payload) {
		t.Fatalf("frame1: kind=%d err=%v payload=%q", kind, err, got)
	}
	kind, got, err = ReadFrame(&buf)
	if err != nil || kind != FrameResize || !bytes.Equal(got, []byte{0, 80, 0, 24}) {
		t.Fatalf("frame2: kind=%d err=%v", kind, err)
	}
	if _, _, err := ReadFrame(&buf); err == nil {
		t.Fatal("流结束应报错")
	}
}

func TestStartPTYEcho(t *testing.T) {
	p, err := Start(80, 24, "/bin/bash -l")
	if err != nil {
		t.Skipf("无 bash 环境: %v", err)
	}
	defer p.Close()

	// 写入命令（早到输入进缓冲, 同步返回是关键行为）
	if _, err := p.Write([]byte("echo pty-echo-42\n")); err != nil {
		t.Fatal(err)
	}
	deadline := 50
	buf := make([]byte, 4096)
	var out bytes.Buffer
	for i := 0; i < 200; i++ {
		n, err := p.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			if bytes.Contains(out.Bytes(), []byte("pty-echo-42")) {
				return // PASS
			}
		}
		if err != nil && err != io.EOF {
			_ = deadline
			return
		}
	}
	t.Fatalf("未读到 echo 输出: %q", out.String())
}

func TestPipeStreams(t *testing.T) {
	// 验证 net.Pipe 上的帧半双工读写（模拟 stream 场景）
	c1, c2 := net.Pipe()
	defer c1.Close()
	defer c2.Close()
	go func() {
		_ = WriteFrame(c1, FrameData, []byte("hello"))
	}()
	kind, payload, err := ReadFrame(c2)
	if err != nil || kind != FrameData || string(payload) != "hello" {
		t.Fatalf("kind=%d payload=%q err=%v", kind, payload, err)
	}
}
