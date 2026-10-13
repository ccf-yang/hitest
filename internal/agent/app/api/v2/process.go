package v2

import (
	"net/http"

	"hitest/internal/agent/app/api/v2/helper"
	"hitest/internal/agent/app/dto/request"
	websocket2 "hitest/internal/agent/utils/websocket"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var wsUpgrade = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// @Tags Process
// @Summary Process ws
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /process/ws [get]
func (b *BaseApi) ProcessWs(c *gin.Context) {
	if !websocket.IsWebSocketUpgrade(c.Request) {
		helper.Success(c)
		return
	}
	ws, err := wsUpgrade.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	wsClient := websocket2.NewWsClient("processClient", ws)
	go wsClient.Read()
	go wsClient.Write()
}

// @Tags Process
// @Summary Stop Process
// @Param request body request.ProcessReq true "request"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /process/stop [post]
// @x-panel-log {"bodyKeys":["PID"],"paramKeys":[],"BeforeFunctions":[],"formatZH":"结束进程 [PID]","formatEN":"结束进程 [PID]"}
func (b *BaseApi) StopProcess(c *gin.Context) {
	var req request.ProcessReq
	if err := helper.CheckBindAndValidate(&req, c); err != nil {
		return
	}
	if err := processService.StopProcess(req); err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.Success(c)
}

// @Tags Process
// @Summary Get Process Info By PID
// @Param pid path int true "PID"
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /process/{pid} [get]
func (b *BaseApi) GetProcessInfoByPID(c *gin.Context) {
	pid, err := helper.GetParamInt32("pid", c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	data, err := processService.GetProcessInfoByPID(pid)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, data)
}

// @Tags Process
// @Summary Get Listening Process
// @Success 200
// @Security ApiKeyAuth
// @Security Timestamp
// @Router /process/listening [post]
func (b *BaseApi) GetListeningProcess(c *gin.Context) {
	procs, err := processService.GetListeningProcess(c)
	if err != nil {
		helper.BadRequest(c, err)
		return
	}
	helper.SuccessWithData(c, procs)
}
