package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"hitest/internal/agent/global"
)

// QuickCommand 预置命令：仅允许执行配置文件中白名单内的命令。
type QuickCommand struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type quickCommandFile struct {
	Commands []QuickCommand `json:"commands" yaml:"commands"`
}

// 快捷命令配置文件路径：<BASE_DIR>/hitest/conf/commands.yaml
func quickCommandFilepath() string {
	return filepath.Join(global.CONF.Base.InstallDir, "hitest", "conf", "commands.yaml")
}

const quickCommandTimeout = 120 * time.Second
const quickCommandOutputLimit = 8 * 1024 // 8KB

// ListQuickCommands 读取预置命令列表；文件不存在时返回空列表。
func (s *SystemService) ListQuickCommands() ([]QuickCommand, error) {
	data, err := os.ReadFile(quickCommandFilepath())
	if err != nil {
		if os.IsNotExist(err) {
			return []QuickCommand{}, nil
		}
		return nil, fmt.Errorf("read commands config failed: %w", err)
	}
	var conf quickCommandFile
	if err := yaml.Unmarshal(data, &conf); err != nil {
		return nil, fmt.Errorf("parse commands config failed: %w", err)
	}
	if conf.Commands == nil {
		conf.Commands = []QuickCommand{}
	}
	return conf.Commands, nil
}

// RunQuickCommand 按名称执行预置命令（白名单校验），返回合并输出（截断至 8KB）。
func (s *SystemService) RunQuickCommand(name string) (string, error) {
	commands, err := s.ListQuickCommands()
	if err != nil {
		return "", err
	}
	var target *QuickCommand
	for i := range commands {
		if commands[i].Name == name {
			target = &commands[i]
			break
		}
	}
	if target == nil {
		return "", fmt.Errorf("quick command not found: %s", name)
	}
	if strings.TrimSpace(target.Command) == "" {
		return "", fmt.Errorf("quick command %s has empty command", name)
	}

	ctx, cancel := context.WithTimeout(context.Background(), quickCommandTimeout)
	defer cancel()

	out, err := runQuickCommandShell(ctx, target.Command)
	if err != nil && len(out) == 0 {
		return "", fmt.Errorf("run quick command failed: %w", err)
	}
	if err != nil {
		out = fmt.Sprintf("%s\n[exit] %v", out, err)
	}
	if len(out) > quickCommandOutputLimit {
		out = out[:quickCommandOutputLimit] + "\n...[truncated]"
	}
	return out, nil
}
