package migration

import (
	"hitest/internal/agent/global"
	"hitest/internal/shared/constant"
	"hitest/internal/shared/model"
)

// Init 极简迁移：仅创建剩余功能所需的表，并写入基础设置。
func Init() {
	if err := global.DB.AutoMigrate(&model.Setting{}); err != nil {
		panic(err)
	}
	if err := global.MonitorDB.AutoMigrate(&model.MonitorBase{}, &model.MonitorNetwork{}, &model.MonitorIO{}); err != nil {
		panic(err)
	}
	if err := global.TaskDB.AutoMigrate(&model.Task{}); err != nil {
		panic(err)
	}
	seedSettings()
	global.LOG.Info("Migration run successfully")
}

func seedSettings() {
	defaults := map[string]string{
		"Language":         "zh",
		"SystemVersion":    global.CONF.Base.Version,
		"SystemStatus":     "Free",
		"MonitorStatus":    constant.StatusDisable,
		"MonitorInterval":  "300",
		"MonitorStoreDays": "30",
		"NtpSite":          "pool.ntp.org",
	}
	for key, value := range defaults {
		if value == "" {
			continue
		}
		var count int64
		global.DB.Model(&model.Setting{}).Where("key = ?", key).Count(&count)
		if count == 0 {
			_ = global.DB.Create(&model.Setting{Key: key, Value: value}).Error
		}
	}
}
