package dir

import (
	"path"

	"hitest/internal/agent/global"
	"hitest/internal/agent/utils/files"
)

func Init() {
	fileOp := files.NewFileOp()
	baseDir := global.CONF.Base.InstallDir
	_, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/docker/compose/"))

	global.Dir.BaseDir, _ = fileOp.CreateDirWithPath(true, baseDir)
	global.Dir.DataDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest"))
	global.Dir.DbDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/db"))
	global.Dir.LogDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/log"))
	global.Dir.TaskDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/log/task"))
	global.Dir.TmpDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/tmp"))

	global.Dir.AppDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/apps"))
	global.Dir.ResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource"))
	global.Dir.IconCacheDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource/icon"))
	global.Dir.AppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource/apps"))
	global.Dir.AppInstallDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/apps"))
	global.Dir.LocalAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource/apps/local"))
	global.Dir.LocalAppInstallDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/apps/local"))
	global.Dir.RemoteAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource/apps/remote"))
	global.Dir.CustomAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource/apps/custom"))
	global.Dir.OfflineAppResourceDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/resource/offline"))
	global.Dir.RuntimeDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/runtime"))
	global.Dir.RecycleBinDir, _ = fileOp.CreateDirWithPath(true, "/.hitest_clash")
	global.Dir.SSLLogDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/log/ssl"))
	global.Dir.McpDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/mcp"))
	global.Dir.ConvertLogDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/log/convert"))
	global.Dir.TensorRTLLMDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/ai/tensorrt_llm"))
	global.Dir.FirewallDir, _ = fileOp.CreateDirWithPath(true, path.Join(baseDir, "hitest/firewall"))
}
