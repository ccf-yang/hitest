package common

import (
	"strings"
	"time"

	"hitest/internal/core/utils/cmd"
)

func LoadTimeZoneByCmd() string {
	loc := time.Now().Location().String()
	if _, err := time.LoadLocation(loc); err != nil {
		loc = "Asia/Shanghai"
	}
	std, err := cmd.NewCommandMgr().RunWithStdout("timedatectl")
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
