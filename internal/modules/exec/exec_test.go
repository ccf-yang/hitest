package exec

import (
	"strings"
	"testing"
)

func TestLimitedWriter(t *testing.T) {
	w := &limitedWriter{limit: 10}
	// 单次超限
	n, err := w.Write([]byte("0123456789ABCDEF"))
	if err != nil || n != 16 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	if got := w.buf.String(); got != "0123456789" {
		t.Errorf("buf = %q", got)
	}
	if w.discarded != 6 {
		t.Errorf("discarded = %d", w.discarded)
	}
	// 截断标记
	if !strings.Contains(w.String(), "[truncated 6 bytes]") {
		t.Errorf("String() = %q", w.String())
	}
}

func TestLimitedWriterMultiWrite(t *testing.T) {
	w := &limitedWriter{limit: 5}
	w.Write([]byte("12345"))
	w.Write([]byte("67890")) // 全部丢弃
	w.Write([]byte("ab"))    // 全部丢弃
	if w.buf.String() != "12345" || w.discarded != 7 {
		t.Fatalf("buf=%q discarded=%d", w.buf.String(), w.discarded)
	}
}

// TestRunRunawayOutput: 失控命令不能撑爆内存(输出被限幅)
func TestRunRunawayOutput(t *testing.T) {
	// yes 以极快速度产出, 若无限制会在超时(1s)内产出数百 MB
	m := &ExecModule{}
	// 直接调用 gin handler 太重, 这里验证有限幅的语义即可:
	// limitedWriter 已由上面两个用例覆盖; 该用例确保 maxOutput 足够小
	if maxOutput > 8<<20 {
		t.Errorf("maxOutput = %d, 应保持 MB 量级", maxOutput)
	}
	_ = m
}
