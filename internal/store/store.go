// Package store: 单一 sqlite 库。
// 起步 2 张表: machines(已注册机器) + kv(杂项设置)。
// 各模块未来需要自己的表时, 在这里登记 AutoMigrate。
package store

import (
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Machine struct {
	ID        string `gorm:"primaryKey"`
	Hostname  string
	OS        string
	FirstSeen time.Time
	LastSeen  time.Time
}

type KV struct {
	K string `gorm:"primaryKey"`
	V string
}

type Store struct {
	db *gorm.DB
}

func Open(dbPath string) (*Store, error) {
	if err := osMkdirAll(filepath.Dir(dbPath)); err != nil {
		return nil, err
	}
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Machine{}, &KV{}); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

// DB 暴露底层 gorm 句柄（模块内联查询用）
func (s *Store) DB() *gorm.DB { return s.db }

// Migrate 登记模块自己的表
func (s *Store) Migrate(models ...any) error {
	return s.db.AutoMigrate(models...)
}

func (s *Store) TouchMachine(id, hostname, goos string) {
	var m Machine
	err := s.db.Where("id = ?", id).First(&m).Error
	now := time.Now()
	if err != nil {
		s.db.Create(&Machine{ID: id, Hostname: hostname, OS: goos, FirstSeen: now, LastSeen: now})
		return
	}
	m.LastSeen = now
	if hostname != "" {
		m.Hostname = hostname
	}
	if goos != "" {
		m.OS = goos
	}
	s.db.Save(&m)
}

func (s *Store) Machines() []Machine {
	var out []Machine
	s.db.Order("last_seen desc").Find(&out)
	return out
}

func (s *Store) KVSet(k, v string) {
	s.db.Save(&KV{K: k, V: v})
}

func (s *Store) KVGet(k string) string {
	var kv KV
	if s.db.Where("k = ?", k).First(&kv).Error == nil {
		return kv.V
	}
	return ""
}
