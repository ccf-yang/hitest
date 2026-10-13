package migration

import (
	"hitest/internal/core/global"
	"hitest/internal/shared/model"
)

// Init 极简迁移：仅创建核心所需的表并写入基础设置。
func Init() {
	if err := global.DB.AutoMigrate(&model.Setting{}, &model.Task{}); err != nil {
		panic(err)
	}
	seedSettings()
	global.LOG.Info("Migration run successfully")
}

func seedSettings() {
	defaults := map[string]string{
		"UserName":              global.CONF.Base.Username,
		"Password":              global.CONF.Base.Password,
		"Language":              global.CONF.Base.Language,
		"SystemVersion":         global.CONF.Base.Version,
		"SystemStatus":          "Free",
		"ServerPort":            global.CONF.Conn.Port,
		"BindAddress":           global.CONF.Conn.BindAddress,
		"SSL":                   global.CONF.Conn.SSL,
		"Ipv6":                  global.CONF.Conn.Ipv6,
		"BindDomain":            "",
		"AllowIPs":              "",
		"AllowIPTrustedProxies": "",
		"NoAuthSetting":         "200",
	}
	for key, value := range defaults {
		var count int64
		global.DB.Model(&model.Setting{}).Where("key = ?", key).Count(&count)
		if count == 0 {
			_ = global.DB.Create(&model.Setting{Key: key, Value: value}).Error
		}
	}
}
