// Package common: 移植自 legacy 的小工具函数
package common

import "fmt"

const (
	b  = 1
	kb = 1024 * b
	mb = 1024 * kb
	gb = 1024 * mb
	tb = 1024 * gb
)

func FormatBytes(bytes uint64) string {
	switch {
	case bytes < kb:
		return fmt.Sprintf("%dB", bytes)
	case bytes < mb:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(kb))
	case bytes < mb*1024:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(mb))
	case bytes < tb:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(gb))
	default:
		return fmt.Sprintf("%.2fTB", float64(bytes)/float64(tb))
	}
}

func LoadSizeUnit2F(value float64) string {
	if value > 1073741824 {
		return fmt.Sprintf("%.2fG", value/1073741824)
	}
	if value > 1048576 {
		return fmt.Sprintf("%.2fM", value/1048576)
	}
	if value > 1024 {
		return fmt.Sprintf("%.2fK", value/1024)
	}
	return fmt.Sprintf("%.2f", value)
}
