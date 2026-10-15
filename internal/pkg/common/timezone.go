package common

import (
	"strings"
	"time"

	cmdx "hitest/internal/pkg/cmdx"
)

// LoadTimeZoneByCmd: 移植自 legacy, 通过 timedatectl 获取系统时区
func LoadTimeZoneByCmd() string {
	loc := time.Now().Location().String()
	if _, err := time.LoadLocation(loc); err != nil {
		loc = "Asia/Shanghai"
	}
	std, err := cmdx.NewCommandMgr(cmdx.WithTimeout(20 * time.Second)).RunWithStdout("timedatectl")
	if err != nil {
		return loc
	}
	for _, line := range strings.Split(std, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 5 || fields[0] != "Time" || fields[1] != "zone:" {
			continue
		}
		if _, err := time.LoadLocation(fields[2]); err != nil {
			return loc
		}
		return fields[2]
	}
	return loc
}
