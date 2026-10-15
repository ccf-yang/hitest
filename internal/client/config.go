// Package client: CLI 客户端（同一二进制的第三种角色）。
// 客户端 = server HTTP API 的又一个消费者:
//   login 存配置 → con(终端WS) / exec(HTTP) / cp(HTTP流式) / ls(HTTP)
package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config 存 ~/.hitest/config.yaml
type Config struct {
	Server string `yaml:"server"` // http(s)://host:port
	Token  string `yaml:"token"`  // access_token（server 开启鉴权时必带）
}

func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".hitest", "config.yaml"), nil
}

func Load() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("尚未登录: 先执行 hitest login <server-url> -t <token>")
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Server == "" {
		return nil, fmt.Errorf("配置缺少 server, 重新执行 hitest login")
	}
	return &cfg, nil
}

func Save(cfg *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// machineURL: 拼 /api/v1/m/{id} 或 local
func (c *Config) machineURL(id string) string {
	if id == "local" {
		return c.Server + "/api/v1/local"
	}
	return fmt.Sprintf("%s/api/v1/m/%s", c.Server, id)
}

func prettyJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
