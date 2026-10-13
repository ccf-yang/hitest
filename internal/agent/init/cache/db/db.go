package badger_db

import (
	"time"

	"github.com/patrickmn/go-cache"
)

type Cache struct {
	db *cache.Cache
}

func NewCacheDB() *Cache {
	db := cache.New(5*time.Minute, 10*time.Minute)
	return &Cache{
		db: db,
	}
}

func (c *Cache) Del(key string) {
	c.db.Delete(key)
}

func (c *Cache) Get(key string) string {
	obj, exist := c.db.Get(key)
	if !exist {
		return ""
	}
	return obj.(string)
}
