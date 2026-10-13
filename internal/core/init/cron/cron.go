package cron

import (
	"time"

	"hitest/internal/core/global"
	"hitest/internal/core/utils/common"

	"github.com/robfig/cron/v3"
)

func Init() {
	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	global.Cron = cron.New(cron.WithLocation(nyc), cron.WithChain(cron.Recover(cron.DefaultLogger)), cron.WithChain(cron.DelayIfStillRunning(cron.DefaultLogger)))
	global.Cron.Start()
}
