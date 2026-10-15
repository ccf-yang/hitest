package disks

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"hitest/internal/store"
)

type DisksModule struct{ svc *DiskService }

func (m *DisksModule) Name() string { return "disks" }

func (m *DisksModule) RegisterRoutes(api gin.IRouter, _ *store.Store) {
	m.svc = &DiskService{}
	api.GET("/disks", m.list)
	api.POST("/disks/partition", m.partition)
	api.POST("/disks/mount", m.mount)
	api.POST("/disks/unmount", m.unmount)
}

func (m *DisksModule) list(c *gin.Context) {
	info, err := m.svc.GetCompleteDiskInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (m *DisksModule) partition(c *gin.Context) {
	var req DiskPartitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := m.svc.PartitionDisk(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (m *DisksModule) mount(c *gin.Context) {
	var req DiskMountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.MountDisk(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *DisksModule) unmount(c *gin.Context) {
	var req DiskUnmountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := m.svc.UnmountDisk(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
