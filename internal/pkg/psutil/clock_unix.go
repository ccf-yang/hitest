//go:build !windows

package psutil

import "github.com/tklauser/go-sysconf"

// clockTicks: /proc/<pid>/stat 里的 starttime 以时钟节拍为单位,
// 节拍数来自 sysconf(_SC_CLK_TCK)（Linux 恒为 100）。
// 取不到时返回默认值并让 ok=false, 调用方沿用 defaultClockTicks。
func clockTicks() (uint64, bool) {
	ticks, err := sysconf.Sysconf(sysconf.SC_CLK_TCK)
	if err != nil || ticks <= 0 {
		return defaultClockTicks, false
	}
	return uint64(ticks), true
}
