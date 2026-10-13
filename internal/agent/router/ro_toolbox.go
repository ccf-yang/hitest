package router

import (
	v2 "hitest/internal/agent/app/api/v2"

	"github.com/gin-gonic/gin"
)

type ToolboxRouter struct{}

func (s *ToolboxRouter) InitRouter(Router *gin.RouterGroup) {
	toolboxRouter := Router.Group("toolbox")
	baseApi := v2.ApiGroupApp.BaseApi
	{
		toolboxRouter.POST("/device/base", baseApi.LoadDeviceBaseInfo)
		toolboxRouter.GET("/device/users", baseApi.LoadUsers)
		toolboxRouter.GET("/device/zone/options", baseApi.LoadTimeOption)
		toolboxRouter.POST("/device/update/conf", baseApi.UpdateDeviceConf)
		toolboxRouter.POST("/device/update/host", baseApi.UpdateDeviceHost)
		toolboxRouter.POST("/device/update/passwd", baseApi.UpdateDevicePasswd)
		toolboxRouter.POST("/device/update/swap", baseApi.UpdateDeviceSwap)
		toolboxRouter.POST("/device/update/byconf", baseApi.UpdateDeviceByFile)
		toolboxRouter.POST("/device/check/dns", baseApi.CheckDNS)
		toolboxRouter.POST("/device/conf", baseApi.LoadDeviceConf)

	}
}
