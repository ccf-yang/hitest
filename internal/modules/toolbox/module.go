package toolbox

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hitest/internal/modules/dto"
	"hitest/internal/store"
)

type ToolboxModule struct{ svc *DeviceService }

func (m *ToolboxModule) Name() string { return "toolbox" }

func (m *ToolboxModule) RegisterRoutes(api gin.IRouter, st *store.Store) {
	m.svc = &DeviceService{st: st}
	api.POST("/toolbox/device/base", m.base)
	api.POST("/toolbox/device/conf", m.conf)
	api.POST("/toolbox/device/update/conf", m.updateConf)   // key/value 形式
	api.POST("/toolbox/device/update/host", m.updateHosts)  // /etc/hosts
	api.POST("/toolbox/device/update/passwd", m.updatePasswd)
	api.POST("/toolbox/device/update/swap", m.updateSwap)
	api.POST("/toolbox/device/update/byconf", m.updateByConf) // 文件内容形式
	api.POST("/toolbox/device/check/dns", m.checkDNS)
	api.GET("/toolbox/device/users", m.users)
	api.GET("/toolbox/device/zone/options", m.zoneOptions)
}

func (m *ToolboxModule) base(c *gin.Context) {
	info, err := m.svc.LoadBaseInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (m *ToolboxModule) conf(c *gin.Context) {
	var req dto.OperationWithName
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	content, err := m.svc.LoadConf(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, content)
}

func (m *ToolboxModule) updateConf(c *gin.Context) {
	var req dto.SettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.Update(req.Key, req.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *ToolboxModule) updateHosts(c *gin.Context) {
	var req []dto.HostHelper
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.UpdateHosts(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *ToolboxModule) updatePasswd(c *gin.Context) {
	var req dto.ChangePasswd
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.UpdatePasswd(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *ToolboxModule) updateSwap(c *gin.Context) {
	var req dto.SwapHelper
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.UpdateSwap(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *ToolboxModule) updateByConf(c *gin.Context) {
	var req dto.UpdateByNameAndFile
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.UpdateByConf(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *ToolboxModule) checkDNS(c *gin.Context) {
	var req struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ok, err := m.svc.CheckDNS(req.Key, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": ok})
}

func (m *ToolboxModule) users(c *gin.Context) {
	users, err := m.svc.LoadUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (m *ToolboxModule) zoneOptions(c *gin.Context) {
	zones, err := m.svc.LoadTimeZone()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, zones)
}
