package viper

import (
	"bytes"
	"fmt"
	"path"
	"strconv"

	"hitest/internal/agent/global"
	"hitest/internal/agent/utils/files"
	"hitest/internal/agent/utils/xpack"
	"hitest/internal/shared/conf"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

func Init() {
	mode := ""
	fileOp := files.NewFileOp()
	v := viper.NewWithOptions()
	v.SetConfigType("yaml")

	config := global.ServerConfig{}
	if err := yaml.Unmarshal(conf.AgentAppYaml, &config); err != nil {
		panic(err)
	}
	if config.Base.Mode != "" {
		mode = config.Base.Mode
	}
	if mode == "dev" && fileOp.Stat("/opt/hitest/conf/app.yaml") {
		v.SetConfigName("app")
		v.AddConfigPath(path.Join("/opt/hitest/conf"))
		if err := v.ReadInConfig(); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	} else {
		reader := bytes.NewReader(conf.AgentAppYaml)
		if err := v.ReadConfig(reader); err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}
	v.OnConfigChange(func(e fsnotify.Event) {
		if err := v.Unmarshal(&global.CONF); err != nil {
			panic(err)
		}
	})
	serverConfig := global.ServerConfig{}
	if err := v.Unmarshal(&serverConfig); err != nil {
		panic(err)
	}

	global.CONF = serverConfig

	initBaseInfo()
	global.Viper = v
}

func initBaseInfo() {
	nodeInfo, err := xpack.MultiNodeProvider.LoadNodeInfo(true)
	if err != nil {
		panic(err)
	}
	global.CONF.Base.InstallDir = nodeInfo.BaseDir
	if !global.IsMaster {
		global.CONF.Base.Port = strconv.FormatUint(uint64(nodeInfo.NodePort), 10)
		if nodeInfo.NodePort == 0 {
			global.CONF.Base.Port = "9999"
		}
	}
}
