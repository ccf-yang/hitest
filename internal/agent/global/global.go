package global

import (
	"context"
	"sync"

	badger_db "hitest/internal/agent/init/cache/db"

	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	DB            *gorm.DB
	MonitorDB     *gorm.DB
	GPUMonitorDB  *gorm.DB
	VLLMMonitorDB *gorm.DB
	TaskDB        *gorm.DB
	CoreDB        *gorm.DB
	AlertDB       *gorm.DB

	LOG   *logrus.Logger
	CONF  ServerConfig
	VALID *validator.Validate
	CACHE *badger_db.Cache
	Viper *viper.Viper

	Dir SystemDir

	Cron              *cron.Cron
	MonitorCronID     cron.EntryID
	VLLMMonitorCronID cron.EntryID
	GPUMonitorCronID  cron.EntryID

	IsMaster bool

	I18n *i18n.Localizer

	AlertBaseJobID     cron.EntryID
	AlertResourceJobID cron.EntryID

	TaskCtxMap = make(map[string]context.CancelFunc)
	taskCtxMu  sync.RWMutex
)

func RegisterTaskCancel(taskID string, cancel context.CancelFunc) {
	taskCtxMu.Lock()
	defer taskCtxMu.Unlock()
	TaskCtxMap[taskID] = cancel
}

func RemoveTaskCancel(taskID string) {
	taskCtxMu.Lock()
	defer taskCtxMu.Unlock()
	delete(TaskCtxMap, taskID)
}

func ResourceURL() string {
	if CONF.Base.IsEnterprise {
		return "https://resource.fit2cloud.com/hitest/resource/v2"
	}
	if CONF.Base.IsFxplay {
		return "https://resource.fit2cloud.com/hitest/resource/v2"
	}
	if CONF.Base.Edition != "intl" {
		return "https://resource.fit2cloud.com/hitest/resource/v2"
	}
	return "https://resource.1panel.pro/v2/resource"
}
