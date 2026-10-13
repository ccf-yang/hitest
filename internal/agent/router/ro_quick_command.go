package router

import (
	"github.com/gin-gonic/gin"

	"hitest/internal/agent/app/api/v2"
)

type QuickCommandRouter struct{}

func (r *QuickCommandRouter) InitRouter(Router *gin.RouterGroup) {
	baseApi := v2.ApiGroupApp.BaseApi
	cmdGroup := Router.Group("commands")
	{
		cmdGroup.GET("/preset", baseApi.ListQuickCommands)
		cmdGroup.POST("/preset/run", baseApi.RunQuickCommand)
	}
}
