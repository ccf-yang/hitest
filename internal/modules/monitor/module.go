package monitor

// MonitorModule: 仪表盘 + 历史监控。
// 移植自 legacy service/dashboard.go + monitor.go。

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"hitest/internal/modules/dto"
	"hitest/internal/store"
)

type MonitorModule struct {
	st *store.Store
	svc *DashboardService
}

func (m *MonitorModule) Name() string { return "monitor" }

func (m *MonitorModule) RegisterRoutes(api gin.IRouter, st *store.Store) {
	m.st = st
	m.svc = &DashboardService{}
	if st != nil {
		_ = st.Migrate(&MonitorBase{}, &MonitorIO{}, &MonitorNetwork{})
		StartSampler(st)
	}

	api.GET("/monitor/base/os", m.osInfo)
	api.GET("/monitor/base/:ioOption/:netOption", m.baseInfo)
	api.GET("/monitor/current/:ioOption/:netOption", m.currentInfo)
	api.GET("/monitor/top/:kind", m.top)
	api.POST("/monitor/restart/:operation", m.restart)
	api.POST("/monitor/search", m.search)
	api.GET("/monitor/setting", m.setting)
	api.POST("/monitor/setting/update", m.updateSetting)
	api.GET("/monitor/iooptions", m.ioOptions)
	api.GET("/monitor/netoptions", m.netOptions)
	api.POST("/monitor/clean", m.clean)
}

func (m *MonitorModule) osInfo(c *gin.Context) {
	info, err := m.svc.LoadOsInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (m *MonitorModule) baseInfo(c *gin.Context) {
	info, err := m.svc.LoadBaseInfo(c.Param("ioOption"), c.Param("netOption"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, info)
}

func (m *MonitorModule) currentInfo(c *gin.Context) {
	c.JSON(http.StatusOK, m.svc.LoadCurrentInfo(c.Param("ioOption"), c.Param("netOption")))
}

func (m *MonitorModule) top(c *gin.Context) {
	switch c.Param("kind") {
	case "cpu":
		c.JSON(http.StatusOK, m.svc.LoadTopCPU())
	case "mem":
		c.JSON(http.StatusOK, m.svc.LoadTopMem())
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "kind must be cpu|mem"})
	}
}

func (m *MonitorModule) restart(c *gin.Context) {
	if err := m.svc.Restart(c.Param("operation")); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------- 历史数据 ----------

func (m *MonitorModule) search(c *gin.Context) {
	if m.st == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no store"})
		return
	}
	var req dto.MonitorSearch
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.StartTime.IsZero() {
		req.StartTime = time.Now().Add(-24 * time.Hour)
	}
	if req.EndTime.IsZero() {
		req.EndTime = time.Now()
	}
	data, err := m.loadMonitorData(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (m *MonitorModule) loadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error) {
	db := m.st.DB().Where("created_at BETWEEN ? AND ?", req.StartTime, req.EndTime)
	var data []dto.MonitorData

	if req.Param == "all" || req.Param == "cpu" || req.Param == "memory" || req.Param == "load" {
		var bases []MonitorBase
		if err := db.Order("created_at").Find(&bases).Error; err != nil {
			return nil, err
		}
		var item dto.MonitorData
		item.Param = "base"
		for _, base := range bases {
			item.Date = append(item.Date, base.CreatedAt)
			if req.Param == "all" || req.Param == "cpu" {
				var processes []dto.Process
				_ = json.Unmarshal([]byte(base.TopCPU), &processes)
				base.TopCPUItems = processes
				base.TopCPU = ""
			}
			if req.Param == "all" || req.Param == "memory" {
				var processes []dto.Process
				_ = json.Unmarshal([]byte(base.TopMem), &processes)
				base.TopMemItems = processes
				base.TopMem = ""
			}
			item.Value = append(item.Value, base)
		}
		data = append(data, item)
	}
	if req.Param == "all" || req.Param == "io" {
		q := m.st.DB().Where("created_at BETWEEN ? AND ?", req.StartTime, req.EndTime)
		if req.IO != "" && req.IO != "all" {
			q = q.Where("name = ?", req.IO)
		}
		var ios []MonitorIO
		if err := q.Order("created_at").Find(&ios).Error; err != nil {
			return nil, err
		}
		var item dto.MonitorData
		item.Param = "io"
		for _, io := range ios {
			item.Date = append(item.Date, io.CreatedAt)
			item.Value = append(item.Value, io)
		}
		data = append(data, item)
	}
	if req.Param == "all" || req.Param == "network" {
		q := m.st.DB().Where("created_at BETWEEN ? AND ?", req.StartTime, req.EndTime)
		if req.Network != "" && req.Network != "all" {
			q = q.Where("name = ?", req.Network)
		}
		var nets []MonitorNetwork
		if err := q.Order("created_at").Find(&nets).Error; err != nil {
			return nil, err
		}
		var item dto.MonitorData
		item.Param = "network"
		for _, n := range nets {
			item.Date = append(item.Date, n.CreatedAt)
			item.Value = append(item.Value, n)
		}
		data = append(data, item)
	}
	return data, nil
}

// ---------- 设置 ----------

func (m *MonitorModule) setting(c *gin.Context) {
	info := dto.MonitorSetting{
		MonitorStatus:  m.st.KVGet("MonitorStatus"),
		MonitorStoreDays: m.st.KVGet("MonitorStoreDays"),
		MonitorInterval:  m.st.KVGet("MonitorInterval"),
		DefaultNetwork:   m.st.KVGet("DefaultNetwork"),
		DefaultIO:        m.st.KVGet("DefaultIO"),
	}
	c.JSON(http.StatusOK, info)
}

func (m *MonitorModule) updateSetting(c *gin.Context) {
	var req dto.MonitorSettingUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	switch req.Key {
	case "MonitorStatus":
		if req.Value != "Enable" && req.Value != "Disable" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid monitoring status"})
			return
		}
	case "MonitorInterval":
		interval, err := strconv.Atoi(req.Value)
		if err != nil || interval < 10 || interval > 43200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "monitoring interval must be between 10 and 43200 seconds"})
			return
		}
	case "MonitorStoreDays":
		days, err := strconv.Atoi(req.Value)
		if err != nil || days < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "monitoring retention must be a positive integer"})
			return
		}
	}
	m.st.KVSet(req.Key, req.Value)
	if req.Key == "MonitorStatus" {
		if req.Value == "Enable" {
			StartSampler(m.st)
		} else {
			stopMonitor()
		}
	}
	if req.Key == "MonitorInterval" && m.st.KVGet("MonitorStatus") == "Enable" {
		stopMonitor()
		StartSampler(m.st)
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *MonitorModule) ioOptions(c *gin.Context) {
	optionSet := map[string]struct{}{}
	if diskStat, err := diskIOCounters(); err == nil {
		for _, item := range diskStat {
			optionSet[item] = struct{}{}
		}
	}
	var names []string
	m.st.DB().Model(&MonitorIO{}).Distinct().Pluck("name", &names)
	for _, name := range names {
		optionSet[name] = struct{}{}
	}
	c.JSON(http.StatusOK, sortedOptions(optionSet))
}

func (m *MonitorModule) netOptions(c *gin.Context) {
	optionSet := map[string]struct{}{}
	for _, item := range netIOCounters() {
		optionSet[item] = struct{}{}
	}
	var names []string
	m.st.DB().Model(&MonitorNetwork{}).Distinct().Pluck("name", &names)
	for _, name := range names {
		optionSet[name] = struct{}{}
	}
	c.JSON(http.StatusOK, sortedOptions(optionSet))
}

func (m *MonitorModule) clean(c *gin.Context) {
	var req dto.MonitorClean
	if err := c.ShouldBindJSON(&req); err != nil || req.Type != "host" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported monitor clean type"})
		return
	}
	db := m.st.DB()
	_ = db.Where("1=1").Delete(&MonitorBase{}).Error
	_ = db.Where("1=1").Delete(&MonitorIO{}).Error
	_ = db.Where("1=1").Delete(&MonitorNetwork{}).Error
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
