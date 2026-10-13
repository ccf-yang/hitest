package cache

import (
	"hitest/internal/agent/global"
	cachedb "hitest/internal/agent/init/cache/db"
)

func Init() {
	global.CACHE = cachedb.NewCacheDB()
}
