package v2

import (
	"errors"

	"hitest/internal/agent/app/api/v2/helper"

	"github.com/gin-gonic/gin"
)

// @Tags Dashboard
// @Summary Load os info
// @Accept json
// @Success 200 {object} dto.OsInfo
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/base/os [get]
func (b *BaseApi) LoadDashboardOsInfo(c *gin.Context) {
	data, err := dashboardService.LoadOsInfo()
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load quick jump options
// @Success 200 {array} dto.QuickJump
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/quick/option [get]
// @Tags Dashboard
// @Summary Update quick jump
// @Accept json
// @Param request body dto.ChangeQuicks true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/quick/change [post]
// @x-panel-log {"bodyKeys":[],"paramKeys":[],"BeforeFunctions":[],"formatZH":"切换快速跳转","formatEN":"change quick jump"}
// @Tags Dashboard
// @Summary Load dashboard base info
// @Accept json
// @Param ioOption path string true "request"
// @Param netOption path string true "request"
// @Success 200 {object} dto.DashboardBase
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/base/:ioOption/:netOption [get]
func (b *BaseApi) LoadDashboardBaseInfo(c *gin.Context) {
	ioOption, ok := c.Params.Get("ioOption")
	if !ok {
		helper.BadRequest(c, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.BadRequest(c, errors.New("error ioOption in path"))
		return
	}
	data, err := dashboardService.LoadBaseInfo(ioOption, netOption)
	if err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load dashboard current info for node
// @Success 200 {object} dto.NodeCurrent
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/current/node [get]
func (b *BaseApi) LoadCurrentInfoForNode(c *gin.Context) {
	data := dashboardService.LoadCurrentInfoForNode()
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load dashboard current info
// @Accept json
// @Param ioOption path string true "request"
// @Param netOption path string true "request"
// @Success 200 {object} dto.DashboardCurrent
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/current/:ioOption/:netOption [get]
func (b *BaseApi) LoadDashboardCurrentInfo(c *gin.Context) {
	ioOption, ok := c.Params.Get("ioOption")
	if !ok {
		helper.BadRequest(c, errors.New("error ioOption in path"))
		return
	}
	netOption, ok := c.Params.Get("netOption")
	if !ok {
		helper.BadRequest(c, errors.New("error netOption in path"))
		return
	}

	data := dashboardService.LoadCurrentInfo(ioOption, netOption)
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load top cpu processes
// @Success 200 {array} dto.Process
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/current/top/cpu [get]
func (b *BaseApi) LoadDashboardTopCPU(c *gin.Context) {
	data := dashboardService.LoadTopCPU()
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary Load top memory processes
// @Success 200 {array} dto.Process
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/current/top/mem [get]
func (b *BaseApi) LoadDashboardTopMem(c *gin.Context) {
	data := dashboardService.LoadTopMem()
	helper.SuccessWithData(c, data)
}

// @Tags Dashboard
// @Summary System restart
// @Accept json
// @Param operation path string true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /dashboard/system/restart/:operation [post]
func (b *BaseApi) SystemRestart(c *gin.Context) {
	operation, ok := c.Params.Get("operation")
	if !ok {
		helper.BadRequest(c, errors.New("error operation in path"))
		return
	}
	if err := dashboardService.Restart(operation); err != nil {
		helper.InternalServer(c, err)
		return
	}
	helper.Success(c)
}
