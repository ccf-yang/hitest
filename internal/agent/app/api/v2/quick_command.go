package v2

import (
	"hitest/internal/agent/app/api/v2/helper"

	"github.com/gin-gonic/gin"
)

// ListQuickCommands GET /commands/preset
func (b *BaseApi) ListQuickCommands(c *gin.Context) {
	commands, err := systemService.ListQuickCommands()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, commands)
}

type runQuickCommandReq struct {
	Name string `json:"name" validate:"required"`
}

// RunQuickCommand POST /commands/preset/run
func (b *BaseApi) RunQuickCommand(c *gin.Context) {
	var req runQuickCommandReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	output, err := systemService.RunQuickCommand(req.Name)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, output)
}
