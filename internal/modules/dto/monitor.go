package dto

import "time"

// ---------- 监控查询/设置（移植自 legacy, 剥离 GPU/VLLM） ----------

type MonitorSearch struct {
	Param     string    `json:"param" validate:"required,oneof=all cpu memory load io network"`
	IO        string    `json:"io"`
	Network   string    `json:"network"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
}

type MonitorData struct {
	Param string        `json:"param"`
	Date  []time.Time   `json:"date"`
	Value []interface{} `json:"value"`
}

type Process struct {
	Name    string  `json:"name"`
	Pid     int32   `json:"pid"`
	Percent float64 `json:"percent"`
	Memory  uint64  `json:"memory"`
	Cmd     string  `json:"cmd"`
	User    string  `json:"user"`
}

type MonitorSetting struct {
	MonitorStatus  string `json:"monitorStatus"`
	MonitorStoreDays string `json:"monitorStoreDays"`
	MonitorInterval  string `json:"monitorInterval"`
	DefaultNetwork   string `json:"defaultNetwork"`
	DefaultIO        string `json:"defaultIO"`
}

type MonitorSettingUpdate struct {
	Key   string `json:"key" validate:"required,oneof=MonitorStatus MonitorStoreDays MonitorInterval DefaultNetwork DefaultIO"`
	Value string `json:"value"`
}

type MonitorClean struct {
	Type string `json:"type" validate:"required,oneof=host"`
}
