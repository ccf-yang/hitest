// Package buserr: legacy 的业务错误 key 机制的极简替代。
// 直接携带中文文案, 不再走 i18n 文件。
package buserr

import (
	"errors"
	"fmt"
)

var zh = map[string]string{
	"ErrCmdIllegal":    "命令包含非法字符",
	"ErrNotExistUser":  "用户不存在",
	"XfsNotFound":      "未找到 xfs_admin / xfs 相关工具",
	"DeviceNotFound":   "设备不存在",
	"DeviceIsMounted":  "设备已挂载",
	"PartitionDiskErr": "磁盘分区失败",
	"FormatDiskErr":    "磁盘格式化失败",
	"MountDiskErr":     "磁盘挂载失败",
	"UnmountDiskErr":   "卸载磁盘失败",
	"ErrCmdTimeout":    "命令超时",
}

func text(key string) string {
	if s, ok := zh[key]; ok {
		return s
	}
	return key
}

func New(key string) error { return errors.New(text(key)) }

func WithName(key, name string) error { return fmt.Errorf("%s: %s", text(key), name) }

func WithErr(key string, err error) error { return fmt.Errorf("%s: %v", text(key), err) }
