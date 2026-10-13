//go:build windows

package service

import (
	"context"
	"os/exec"
)

func runQuickCommandShell(ctx context.Context, command string) (string, error) {
	cmd := exec.CommandContext(ctx, "cmd", "/C", command)
	out, err := cmd.CombinedOutput()
	return string(out), err
}
