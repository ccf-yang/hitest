package hook

import (
	"time"

	"hitest/internal/agent/app/repo"
	"hitest/internal/agent/global"
	"hitest/internal/agent/utils/xpack"
	"hitest/internal/shared/model"

	"gorm.io/gorm"
)

func Init() {
	global.LOG.Info("agent hook: init global data")
	initGlobalData()
	global.LOG.Info("agent hook: init monitor db")
	initMonitorDB()
}

func initGlobalData() {
	settingRepo := repo.NewISettingRepo()
	if _, err := settingRepo.GetValueByKey("SystemStatus"); err != nil {
		_ = settingRepo.Create("SystemStatus", "Free")
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}
	node, _ := xpack.MultiNodeProvider.LoadNodeInfo(false)
	if len(node.Version) != 0 {
		_ = settingRepo.Update("SystemVersion", node.Version)
	}
	global.CONF.Base.Version = node.Version
	global.CONF.Base.Edition, _ = settingRepo.GetValueByKey("Edition")
	global.CONF.Base.EncryptKey, _ = settingRepo.GetValueByKey("EncryptKey")
}

func initMonitorDB() {
	_ = global.MonitorDB.AutoMigrate(&model.MonitorBase{}, &model.MonitorNetwork{}, &model.MonitorIO{})
	_ = global.TaskDB.AutoMigrate(&model.Task{})
	go ensureMonitorIndexes()
}

func ensureMonitorIndexes() {
	indexes := []struct {
		db   *gorm.DB
		stmt string
	}{
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_bases_created ON monitor_bases(created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_ios_created ON monitor_ios(created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_networks_created ON monitor_networks(created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_ios_name_created ON monitor_ios(name, created_at)"},
		{global.MonitorDB, "CREATE INDEX IF NOT EXISTS idx_monitor_networks_name_created ON monitor_networks(name, created_at)"},
	}
	start := time.Now()
	for _, index := range indexes {
		if err := index.db.Exec(index.stmt).Error; err != nil {
			global.LOG.Warnf("create monitor index failed, stmt: %s, err: %v", index.stmt, err)
		}
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		global.LOG.Infof("monitor indexes ready, took %s", elapsed)
	}
}
