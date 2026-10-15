package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

// LoadConfig 读取 serve 配置; 文件不存在时生成默认配置（含随机 token）。
func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := &Config{Port: 8080, Token: randomToken()}
		if err := save(path, cfg); err != nil {
			return nil, err
		}
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Token == "" {
		cfg.Token = randomToken()
		if err := save(path, &cfg); err != nil {
			return nil, err
		}
	}
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	return &cfg, nil
}

func save(path string, cfg *Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
