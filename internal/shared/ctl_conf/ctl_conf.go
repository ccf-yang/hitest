package ctl_conf

import (
	"fmt"
	"os"
	"strings"
)

const defaultFile = "/usr/local/bin/hitestctl"

// baseDirPointer 是首次初始化后写入的「安装目录指针」，
// 供后续启动在没有 hitestctl / 环境变量时定位 BASE_DIR。
const baseDirPointer = "/etc/hitest/base_dir"

// defaultValues 用于在没有 hitestctl 文件、也没有环境变量时兜底，避免裸二进制启动 panic。
var defaultValues = map[string]string{
	"BASE_DIR":          "/opt",
	"ORIGINAL_PORT":     "9999",
	"ORIGINAL_VERSION":  "v2.0.0",
	"ORIGINAL_USERNAME": "admin",
	"ORIGINAL_PASSWORD": "admin123",
	"ORIGINAL_ENTRANCE": "",
	"LANGUAGE":          "zh",
	"PANEL_EDITION":     "",
	"CHANGE_USER_INFO":  "",
}

// lookup 按 环境变量(HITEST_<KEY>) -> hitestctl -> BASE_DIR 指针文件的顺序查找。
func lookup(key string) (string, bool) {
	if value := strings.TrimSpace(os.Getenv("HITEST_" + key)); value != "" {
		return value, true
	}
	if value, err := LoadFromFile(defaultFile, key); err == nil {
		return value, true
	}
	if key == "BASE_DIR" {
		if data, err := os.ReadFile(baseDirPointer); err == nil {
			if value := strings.TrimSpace(string(data)); value != "" {
				return value, true
			}
		}
	}
	return "", false
}

// Load 返回配置值：环境变量 > hitestctl > 内置默认值；不再 panic。
func Load(key string) string {
	if value, ok := lookup(key); ok && value != "" && value != `""` {
		return value
	}
	if fallback, ok := defaultValues[key]; ok {
		return fallback
	}
	return ""
}

// LoadWithoutPanic 返回配置值，找不到时返回空串（不套用内置默认值）。
func LoadWithoutPanic(key string) string {
	value, _ := lookup(key)
	return value
}

func LoadFromFile(filePath, key string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	prefix := key + "="
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix)), nil
		}
	}
	return "", fmt.Errorf("error `%s` find in %s", key, filePath)
}

func UpdateInFile(filePath, key, value string) error {
	return rewriteFile(filePath, func(line string) (string, bool) {
		if strings.HasPrefix(line, key+"=") {
			return key + "=" + value, true
		}
		return line, true
	})
}

func RemoveValueFromFile(filePath, key, value string) error {
	target := key + "=" + value
	return rewriteFile(filePath, func(line string) (string, bool) {
		return line, line != target
	})
}

func rewriteFile(filePath string, rewrite func(string) (string, bool)) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	content := rewriteLines(string(data), rewrite)
	return os.WriteFile(filePath, []byte(content), info.Mode().Perm())
}

func rewriteLines(data string, rewrite func(string) (string, bool)) string {
	hasTrailingNewline := strings.HasSuffix(data, "\n")
	lines := strings.Split(strings.TrimSuffix(data, "\n"), "\n")
	rewritten := make([]string, 0, len(lines))
	for _, line := range lines {
		newLine, keep := rewrite(line)
		if keep {
			rewritten = append(rewritten, newLine)
		}
	}
	content := strings.Join(rewritten, "\n")
	if hasTrailingNewline {
		content += "\n"
	}
	return content
}
