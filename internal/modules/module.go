// Package modules 是功能层：一目录一功能，功能 = 普通 HTTP handler。
// 传输无关：模块代码不知道请求来自本地还是远程隧道。
// 加新功能 = 加一个目录 + 在 All() 里加一行。
package modules

import (
	"github.com/gin-gonic/gin"

	"hitest/internal/modules/commands"
	"hitest/internal/modules/disks"
	"hitest/internal/modules/exec"
	"hitest/internal/modules/files"
	"hitest/internal/modules/monitor"
	"hitest/internal/modules/process"
	"hitest/internal/modules/toolbox"
	"hitest/internal/store"
)

// Module 是所有功能模块的接口
type Module interface {
	Name() string
	RegisterRoutes(api gin.IRouter, st *store.Store)
}

// All 返回启用的功能模块（注册表：加功能在这里加一行）
func All() []Module {
	return []Module{
		&monitor.MonitorModule{},
		&files.FilesModule{},
		&exec.ExecModule{},
		&disks.DisksModule{},
		&toolbox.ToolboxModule{},
		&process.ProcessModule{},
		&commands.CommandsModule{},
	}
}
