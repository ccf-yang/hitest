package hook

import (
	"strings"

	"hitest/internal/core/app/repo"
	"hitest/internal/core/global"
	"hitest/internal/core/utils/common"
	"hitest/internal/core/utils/encrypt"
	"hitest/internal/shared/ctl_conf"
)

func Init() {
	settingRepo := repo.NewISettingRepo()
	storedVersion, _ := settingRepo.GetValueByKey("SystemVersion")
	if storedVersion != global.CONF.Base.Version {
		if err := settingRepo.Update("SystemVersion", global.CONF.Base.Version); err != nil {
			global.LOG.Fatalf("sync system version before start failed, err: %v", err)
		}
		global.LOG.Infof("sync system version from installed package: %s -> %s", storedVersion, global.CONF.Base.Version)
	}
	if v, _ := settingRepo.GetValueByKey("ServerPort"); v != "" {
		global.CONF.Conn.Port = v
	}
	if v, _ := settingRepo.GetValueByKey("Ipv6"); v != "" {
		global.CONF.Conn.Ipv6 = v
	}
	if v, _ := settingRepo.GetValueByKey("Edition"); v != "" {
		global.CONF.Base.Edition = v
	}
	if v, _ := settingRepo.GetValueByKey("BindAddress"); v != "" {
		global.CONF.Conn.BindAddress = v
	}
	if v, _ := settingRepo.GetValueByKey("SSL"); v != "" {
		global.CONF.Conn.SSL = v
	}
	if err := settingRepo.Update("SystemStatus", "Free"); err != nil {
		global.LOG.Fatalf("init service before start failed, err: %v", err)
	}

	handleUserInfo(global.CONF.Base.ChangeUserInfo, settingRepo)
}

func handleUserInfo(tags string, settingRepo repo.ISettingRepo) {
	if len(tags) == 0 {
		return
	}
	settingMap := make(map[string]string)
	if tags == "use_existing" {
		settingMap["ServerPort"] = ctl_conf.Load("ORIGINAL_PORT")
		global.CONF.Conn.Port = settingMap["ServerPort"]
		settingMap["UserName"] = global.CONF.Base.Username
		settingMap["Password"] = global.CONF.Base.Password
		settingMap["SecurityEntrance"] = global.CONF.Conn.Entrance
		settingMap["SystemVersion"] = ctl_conf.Load("ORIGINAL_VERSION")
		global.CONF.Base.Version = settingMap["SystemVersion"]
		settingMap["Language"] = global.CONF.Base.Language
	}
	if tags == "all" {
		settingMap["UserName"] = common.RandStrAndNum(10)
		settingMap["Password"] = common.RandStrAndNum(10)
		settingMap["SecurityEntrance"] = common.RandStrAndNum(10)
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "username") {
		settingMap["UserName"] = common.RandStrAndNum(10)
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "password") {
		settingMap["Password"] = common.RandStrAndNum(10)
	}
	if strings.Contains(global.CONF.Base.ChangeUserInfo, "entrance") {
		settingMap["SecurityEntrance"] = common.RandStrAndNum(10)
	}
	for key, val := range settingMap {
		if len(val) == 0 {
			continue
		}
		if key == "Password" {
			val, _ = encrypt.StringEncrypt(val)
		}
		if err := settingRepo.Update(key, val); err != nil {
			global.LOG.Errorf("update %s before start failed, err: %v", key, err)
		}
	}

	_ = ctl_conf.RemoveValueFromFile("/usr/local/bin/hitestctl", "CHANGE_USER_INFO", global.CONF.Base.ChangeUserInfo)
	_ = ctl_conf.UpdateInFile("/usr/local/bin/hitestctl", "ORIGINAL_PASSWORD", "**********")
}
