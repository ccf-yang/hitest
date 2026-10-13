package service

import "hitest/internal/agent/app/repo"

var (
	monitorRepo = repo.NewIMonitorRepo()
	settingRepo = repo.NewISettingRepo()
	taskRepo    = repo.NewITaskRepo()
)
