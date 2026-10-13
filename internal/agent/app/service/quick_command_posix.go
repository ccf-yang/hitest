//go:build !windows

package service

import (
	"context"
	"os/exec"
)

// runQuickCommandShell 通过 /bin/sh -c 执行预置命令（白名单内的整行命令）。
func runQuickCommandShell(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
