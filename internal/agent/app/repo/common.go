package repo

import (
	"time"

	"gorm.io/gorm"
)

type DBOption func(*gorm.DB) *gorm.DB

func WithByName(name string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("name = ?", name)
	}
}

func WithByCreatedAt(startTime, endTime time.Time) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("created_at > ? AND created_at < ?", startTime, endTime)
	}
}
