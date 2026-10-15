//go:build windows

package commands

import (
	"context"
	"os"
	"os/exec"
)

// runQuickCommandShell 通过 %ComSpec% /c 执行预置命令（白名单内的整行命令）。
func runQuickCommandShell(ctx context.Context, command string) (string, error) {
	shell := os.Getenv("ComSpec")
	if shell == "" {
		shell = "cmd"
	}
	cmd := exec.CommandContext(ctx, shell, "/c", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
