//go:build !windows

package cmd

import (
	"os/exec"
	"syscall"
)

// sysProcAttr: 让子进程自成一个进程组, 便于超时时整组回收。
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}

// killTree: 杀掉整个进程组（负 PID = 进程组）。
func killTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil || cmd.Process.Pid <= 0 {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
