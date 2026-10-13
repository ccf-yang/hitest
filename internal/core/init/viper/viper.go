package viper

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"hitest/internal/core/global"
	"hitest/internal/shared/conf"
	"hitest/internal/shared/ctl_conf"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Init 读取配置的优先级：
//  1. 内置默认配置（conf/app.yaml）
//  2. 配置文件 <BASE_DIR>/hitest/conf/app.yaml（首次运行由 web 初始化向导生成）
//  3. 环境变量（HITEST_BASE_DIR / HITEST_PORT ...）
//
// 不再依赖 /usr/local/bin/hitestctl，缺失时使用默认值而非 panic。
func Init() {
	baseDir := ctl_conf.Load("BASE_DIR")
	if baseDir == "" {
		baseDir = "/opt"
	}

	v := viper.NewWithOptions()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(bytes.NewReader(conf.CoreAppYaml)); err != nil {
		panic(fmt.Errorf("read default config failed: %w", err))
	}

	confFile := path.Join(baseDir, "hitest", "conf", "app.yaml")
	if _, err := os.Stat(confFile); err == nil {
		v.SetConfigFile(confFile)
		if err := v.MergeInConfig(); err != nil {
			panic(fmt.Errorf("read config file %s failed: %w", confFile, err))
		}
		v.OnConfigChange(func(e fsnotify.Event) {
			if err := v.Unmarshal(&global.CONF); err != nil {
				panic(err)
			}
		})
		v.WatchConfig()
	}

	if port := os.Getenv("HITEST_PORT"); port != "" {
		v.Set("conn.port", port)
	} else if port := ctl_conf.LoadWithoutPanic("ORIGINAL_PORT"); port != "" {
		v.Set("conn.port", port)
	}

	serverConfig := global.ServerConfig{}
	if err := v.Unmarshal(&serverConfig); err != nil {
		panic(fmt.Errorf("unmarshal config failed: %w", err))
	}

	if serverConfig.Base.InstallDir == "" {
		serverConfig.Base.InstallDir = baseDir
	}
	if serverConfig.Conn.Port == "" {
		serverConfig.Conn.Port = "9999"
	}
	if serverConfig.Base.Mode == "" {
		serverConfig.Base.Mode = "dev"
	}
	if serverConfig.Base.Version == "" {
		serverConfig.Base.Version = "v2.0.0"
	}
	if serverConfig.Base.Username == "" {
		serverConfig.Base.Username = "admin"
	}
	if serverConfig.Base.Password == "" {
		serverConfig.Base.Password = "admin123"
	}
	if serverConfig.Base.Language == "" {
		serverConfig.Base.Language = "zh"
	}
	if serverConfig.Base.Edition == "intl" {
		serverConfig.Base.Edition = "intl"
	} else {
		serverConfig.Base.Edition = "cn"
	}
	serverConfig.Base.ChangeUserInfo = loadChangeInfo()

	global.CONF = serverConfig
	global.Viper = v
}

func loadChangeInfo() string {
	return ctl_conf.LoadWithoutPanic("CHANGE_USER_INFO")
}
