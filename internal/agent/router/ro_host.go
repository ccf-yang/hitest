package router

import (
	v2 "hitest/internal/agent/app/api/v2"

	"github.com/gin-gonic/gin"
)

type HostRouter struct{}

func (s *HostRouter) InitRouter(Router *gin.RouterGroup) {
	hostRouter := Router.Group("hosts")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		hostRouter.POST("/monitor/search", baseApi.LoadMonitor)
		hostRouter.POST("/monitor/clean", baseApi.CleanMonitor)
		hostRouter.GET("/monitor/netoptions", baseApi.GetNetworkOptions)
		hostRouter.GET("/monitor/iooptions", baseApi.GetIOOptions)
		hostRouter.GET("/monitor/setting", baseApi.LoadMonitorSetting)
		hostRouter.POST("/monitor/setting/update", baseApi.UpdateMonitorSetting)

		hostRouter.GET("/disks", baseApi.GetCompleteDiskInfo)
		hostRouter.POST("/disks/partition", baseApi.PartitionDisk)
		hostRouter.POST("/disks/mount", baseApi.MountDisk)
		hostRouter.POST("/disks/unmount", baseApi.UnmountDisk)

		hostRouter.GET("/components/:name", baseApi.CheckComponentExistence)
	}
}
