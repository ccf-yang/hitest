package db

import (
	"path"

	"hitest/internal/core/global"
	"hitest/internal/core/utils/common"
)

func Init() {
	global.DB = common.LoadDBConnByPath(path.Join(global.CONF.Base.InstallDir, "hitest/db/core.db"), "core")
	global.TaskDB = common.LoadDBConnByPath(path.Join(global.CONF.Base.InstallDir, "hitest/db/task.db"), "task")
	global.AgentDB = common.LoadDBConnByPath(path.Join(global.CONF.Base.InstallDir, "hitest/db/agent.db"), "agent")
	global.AlertDB = common.LoadDBConnByPath(path.Join(global.CONF.Base.InstallDir, "hitest/db/alert.db"), "alert")
}
