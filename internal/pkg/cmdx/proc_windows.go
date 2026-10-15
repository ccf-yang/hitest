//go:build windows

package cmd

import (
	"os/exec"
	"strconv"
	"syscall"
)

// createNewProcessGroup: 与 unix 的 Setpgid 对应的进程分组标志。
const createNewProcessGroup = 0x00000200

// sysProcAttr: 让子进程自成一个进程组, 超时时可整组回收。
func sysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: createNewProcessGroup}
}

// killTree: 用 taskkill /T /F 连同子进程一起结束（Windows 没有负 PID 进程组语义）。
// 只尽力而为, 调用方仍需 Process.Kill() 收尾。
func killTree(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil || cmd.Process.Pid <= 0 {
		return
	}
	killer := exec.Command("taskkill", "/T", "/F", "/PID", strconv.Itoa(cmd.Process.Pid))
	killer.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = killer.Run()
}
