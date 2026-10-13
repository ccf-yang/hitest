package session

import (
	"hitest/internal/core/global"
	"hitest/internal/core/init/session/psession"
)

func Init() {
	global.SESSION = psession.NewPSession("")
	global.LOG.Info("init in-memory session successfully")
}
