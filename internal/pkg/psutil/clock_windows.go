//go:build windows

package psutil

// clockTicks: Windows 上没有 /proc, 基于 procfs 的进程启动时间兜底不会被走到,
// 沿用 defaultClockTicks。
func clockTicks() (uint64, bool) { return defaultClockTicks, false }
