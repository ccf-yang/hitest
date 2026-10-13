package middleware

import (
	"hitest/internal/core/app/api/v2/helper"
	"hitest/internal/core/app/repo"

	"github.com/gin-gonic/gin"
)

func GlobalLoading() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.GetValueByKey("SystemStatus")
		if err != nil {
			// 未初始化该设置时视为可用，避免首次启动被中间件阻断。
			c.Next()
			return
		}
		if status != "Free" {
			helper.ErrorWithDetail(c, 407, status, err)
			return
		}
		c.Next()
	}
}
