package router

import (
	v2 "hitest/internal/agent/app/api/v2"
	"hitest/internal/agent/global"
	"hitest/internal/agent/i18n"
	"hitest/internal/agent/middleware"
	rou "hitest/internal/agent/router"

	"github.com/gin-gonic/gin"
)

var (
	Router *gin.Engine
)

func Routers() *gin.Engine {
	Router = gin.Default()
	Router.Use(i18n.UseI18n())

	PrivateGroup := Router.Group("/api/v2")
	if !global.IsMaster {
		PrivateGroup.Use(middleware.Certificate())
	}
	PrivateGroup.Use(middleware.OperationResolveMeta())
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}
	PrivateGroup.GET("/health/check", v2.ApiGroupApp.BaseApi.CheckHealth)

	return Router
}
