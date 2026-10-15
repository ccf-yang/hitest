package process

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"hitest/internal/store"
)

type ProcessModule struct{ svc *ProcessService }

func (m *ProcessModule) Name() string { return "process" }

func (m *ProcessModule) RegisterRoutes(api gin.IRouter, _ *store.Store) {
	m.svc = &ProcessService{}
	api.GET("/process/:pid", m.info)
	api.POST("/process/stop", m.stop)
	api.POST("/process/listening", m.listening)
	// TODO: /process/ws 实时推送（legacy WsClient 框架, 接 Flutter 时移植）
}

func (m *ProcessModule) info(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("pid"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid pid"})
		return
	}
	data, err := m.svc.GetProcessInfoByPID(int32(pid))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (m *ProcessModule) stop(c *gin.Context) {
	var req ProcessReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.StopProcess(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *ProcessModule) listening(c *gin.Context) {
	data, err := m.svc.GetListeningProcess(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}
