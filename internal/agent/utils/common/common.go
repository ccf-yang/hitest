package common

import (
	"fmt"
	"strings"
	"time"

	"hitest/internal/agent/utils/cmd"
	"hitest/internal/shared/ctl_conf"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func LoadSizeUnit2F(value float64) string {
	if value > 1073741824 {
		return fmt.Sprintf("%.2fG", value/1073741824)
	}
	if value > 1048576 {
		return fmt.Sprintf("%.2fM", value/1048576)
	}
	if value > 1024 {
		return fmt.Sprintf("%.2fK", value/1024)
	}
	return fmt.Sprintf("%.2f", value)
}

func LoadTimeZoneByCmd() string {
	loc := time.Now().Location().String()
	if _, err := time.LoadLocation(loc); err != nil {
		loc = "Asia/Shanghai"
	}
	std, err := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Second)).RunWithStdout("timedatectl")
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

// ParseIPLoose parses an IP address, accepting bracketed IPv6 forms
// such as "[::1]" or "[fe80::1%eth0]" in addition to the bare forms
// that net.ParseIP supports natively. It also trims surrounding
// whitespace. Returns nil for non-IP input, matching net.ParseIP.
//
// This is required because some flows pass the host portion of a URL
// (e.g. "[::1]") rather than a bare IP address. Rejecting that form
// caused hitest-dev/hitest#12646 — the panel SSL self-sign workflow
// reported “domain format invalid” for IPv6 hosts.

const (
	b  = uint64(1)
	kb = 1024 * b
	mb = 1024 * kb
	gb = 1024 * mb
)

func FormatBytes(bytes uint64) string {
	switch {
	case bytes < kb:
		return fmt.Sprintf("%dB", bytes)
	case bytes < mb:
		return fmt.Sprintf("%.2fKB", float64(bytes)/float64(kb))
	case bytes < gb:
		return fmt.Sprintf("%.2fMB", float64(bytes)/float64(mb))
	default:
		return fmt.Sprintf("%.2fGB", float64(bytes)/float64(gb))
	}
}

func LoadParams(param string) string {
	return ctl_conf.Load(param)
}
