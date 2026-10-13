package v2

import "hitest/internal/agent/app/service"

type ApiGroup struct {
	BaseApi
}

var ApiGroupApp = new(ApiGroup)

type BaseApi struct{}

var (
	dashboardService = service.NewIDashboardService()
	monitorService   = service.NewIMonitorService()
	systemService    = service.NewISystemService()
	deviceService    = service.NewIDeviceService()
	processService   = service.NewIProcessService()
	diskService      = service.NewIDiskService()
)
