package process

import (
	"github.com/shirou/gopsutil/v4/net"

	"github.com/shirou/gopsutil/v4/process"
)

// 移植自 legacy utils/websocket/process_data.go（只留数据结构）
type PsProcessData struct {
	PID            int32  `json:"PID"`
	Name           string `json:"name"`
	PPID           int32  `json:"PPID"`
	Username       string `json:"username"`
	Status         string `json:"status"`
	StartTime      string `json:"startTime"`
	NumThreads     int32  `json:"numThreads"`
	NumConnections int    `json:"numConnections"`
	CpuPercent     string `json:"cpuPercent"`

	DiskRead  string `json:"diskRead"`
	DiskWrite string `json:"diskWrite"`
	CmdLine   string `json:"cmdLine"`

	Rss    string `json:"rss"`
	VMS    string `json:"vms"`
	HWM    string `json:"hwm"`
	Data   string `json:"data"`
	Stack  string `json:"stack"`
	Locked string `json:"locked"`
	Swap   string `json:"swap"`
	Dirty  string `json:"dirty"`
	PSS    string `json:"pss"`
	USS    string `json:"uss"`
	Shared string `json:"shared"`
	Text   string `json:"text"`

	CpuValue float64 `json:"cpuValue"`
	RssValue uint64  `json:"rssValue"`

	Envs []string `json:"envs"`

	OpenFiles []process.OpenFilesStat `json:"openFiles"`
	Connects  []ProcessConnect        `json:"connects"`
}
type ProcessConnect struct {
	Type   string   `json:"type"`
	Status string   `json:"status"`
	Laddr  net.Addr `json:"localaddr"`
	Raddr  net.Addr `json:"remoteaddr"`
	PID    int32    `json:"PID"`
	Name   string   `json:"name"`
}
type ProcessConnects []ProcessConnect


type ProcessReq struct {
	PID int32 `json:"PID" validate:"required"`
}
