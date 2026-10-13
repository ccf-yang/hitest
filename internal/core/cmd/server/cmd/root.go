package cmd

import (
	"fmt"
	"os/user"
	"path"
	"time"

	"hitest/internal/core/server"
	"hitest/internal/core/utils/common"
	"hitest/internal/shared/ctl_conf"

	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var language string

func init() {
	RootCmd.PersistentFlags().StringVarP(&language, "language", "l", "en", "Set the language")
}

var RootCmd = &cobra.Command{
	Use: "hitest",
	RunE: func(cmd *cobra.Command, args []string) error {
		server.Start()
		return nil
	},
}

type setting struct {
	ID        uint      `gorm:"primarykey;AUTO_INCREMENT" json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Key       string    `json:"key" gorm:"type:varchar(256);not null;"`
	Value     string    `json:"value" gorm:"type:varchar(256)"`
	About     string    `json:"about" gorm:"type:longText"`
}

func loadDBConn(dbName string) (*gorm.DB, error) {
	baseDir, err := loadBaseDir()
	if err != nil {
		return nil, err
	}

	db, err := common.GetDBWithPath(path.Join(baseDir, "hitest/db", dbName))
	if err != nil {
		return nil, fmt.Errorf("init my db conn failed, err: %v", err)
	}
	return db, nil
}

func loadBaseDir() (string, error) {
	baseDir := ctl_conf.Load("BASE_DIR")
	if len(baseDir) == 0 {
		return "", fmt.Errorf("error `BASE_DIR` not found")
	}
	return baseDir, nil
}

func getSettingByKey(db *gorm.DB, key string) string {
	var setting setting
	_ = db.Where("key = ?", key).First(&setting).Error
	return setting.Value
}

type LoginLog struct{}

func setSettingByKey(db *gorm.DB, key, value string) error {
	return db.Model(&setting{}).Where("key = ?", key).Updates(map[string]interface{}{"value": value}).Error
}

func isRoot() bool {
	currentUser, err := user.Current()
	if err != nil {
		return false
	}
	return currentUser.Uid == "0"
}
