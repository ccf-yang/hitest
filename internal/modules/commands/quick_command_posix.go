//go:build !windows

package commands

import (
	"context"
	"log"
	"os/exec"
)

// runQuickCommandShell 通过 /bin/sh -c 执行预置命令（白名单内的整行命令）。
func runQuickCommandShell(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	out := &cappedWriter{limit: inFlightOutputLimit}
	cmd.Stdout = out
	cmd.Stderr = out
	err := cmd.Run()
	if n := out.Truncated(); n > 0 {
		log.Printf("[commands] 预置命令输出超内存上限, 已丢弃 %d 字节", n)
	}
	return out.String(), err
}
