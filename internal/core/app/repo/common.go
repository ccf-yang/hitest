package repo

import (
	"hitest/internal/core/global"

	"gorm.io/gorm"
)

func WithByKey(key string) global.DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("key = ?", key)
	}
}
