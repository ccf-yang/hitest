package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"hitest/internal/agent/app/repo"
	"hitest/internal/agent/buserr"
	"hitest/internal/shared/constant"

	"hitest/internal/agent/app/dto"
	"hitest/internal/agent/global"
	"hitest/internal/agent/utils/common"
	"hitest/internal/agent/utils/psutil"
	"hitest/internal/shared/model"

	"github.com/robfig/cron/v3"
	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

type MonitorService struct {
	DiskIO chan ([]disk.IOCountersStat)
	NetIO  chan ([]net.IOCountersStat)
	ctx    context.Context
}

var (
	monitorSettingMutex sync.Mutex
	monitorCancel       context.CancelFunc
	hostSysPath         = loadHostSysPath()

	blockDevicePartitionCache sync.Map
)

type IMonitorService interface {
	Run()
	LoadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error)
	LoadSetting() (*dto.MonitorSetting, error)
	UpdateSetting(key, value string) error
	CleanData(monitorType string) error

	LoadIOOptions() []string
	LoadNetworkOptions() []string

	saveIODataToDB(ctx context.Context, interval float64)
	saveNetDataToDB(ctx context.Context, interval float64)
}

func NewIMonitorService() IMonitorService {
	return &MonitorService{
		DiskIO: make(chan []disk.IOCountersStat, 2),
		NetIO:  make(chan []net.IOCountersStat, 2),
		ctx:    context.Background(),
	}
}

func (m *MonitorService) LoadMonitorData(req dto.MonitorSearch) ([]dto.MonitorData, error) {
	loc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	req.StartTime = req.StartTime.In(loc)
	req.EndTime = req.EndTime.In(loc)

	var data []dto.MonitorData
	if req.Param == "all" || req.Param == "cpu" || req.Param == "memory" || req.Param == "load" {
		bases, err := monitorRepo.GetBase(repo.WithByCreatedAt(req.StartTime, req.EndTime))
		if err != nil {
			return nil, err
		}

		var itemData dto.MonitorData
		itemData.Param = "base"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			if req.Param == "all" || req.Param == "cpu" {
				var processes []dto.Process
				_ = json.Unmarshal([]byte(base.TopCPU), &processes)
				base.TopCPUItems = processes
				base.TopCPU = ""
			}
			if req.Param == "all" || req.Param == "memory" {
				var processes []dto.Process
				_ = json.Unmarshal([]byte(base.TopMem), &processes)
				base.TopMemItems = processes
				base.TopMem = ""
			}
			itemData.Value = append(itemData.Value, base)
		}
		data = append(data, itemData)
	}
	if req.Param == "all" || req.Param == "io" {
		ioOpts := []repo.DBOption{repo.WithByCreatedAt(req.StartTime, req.EndTime)}
		if len(req.IO) != 0 {
			ioOpts = append(ioOpts, repo.WithByName(req.IO))
		}
		bases, err := monitorRepo.GetIO(ioOpts...)
		if err != nil {
			return nil, err
		}

		var itemData dto.MonitorData
		itemData.Param = "io"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		data = append(data, itemData)
	}
	if req.Param == "all" || req.Param == "network" {
		netOpts := []repo.DBOption{repo.WithByCreatedAt(req.StartTime, req.EndTime)}
		if len(req.Network) != 0 {
			netOpts = append(netOpts, repo.WithByName(req.Network))
		}
		bases, err := monitorRepo.GetNetwork(netOpts...)
		if err != nil {
			return nil, err
		}

		var itemData dto.MonitorData
		itemData.Param = "network"
		for _, base := range bases {
			itemData.Date = append(itemData.Date, base.CreatedAt)
			itemData.Value = append(itemData.Value, base)
		}
		data = append(data, itemData)
	}
	return data, nil
}

func (m *MonitorService) LoadIOOptions() []string {
	optionSet := make(map[string]struct{})
	if diskStat, err := disk.IOCounters(); err == nil {
		for _, item := range diskStat {
			optionSet[item.Name] = struct{}{}
		}
	}
	// union with names recorded in the monitor db so removed devices stay selectable
	if names, err := monitorRepo.GetIONames(); err == nil {
		for _, name := range names {
			optionSet[name] = struct{}{}
		}
	}
	return sortedMonitorOptions(optionSet)
}

func (m *MonitorService) LoadNetworkOptions() []string {
	optionSet := make(map[string]struct{})
	if netStat, err := net.IOCounters(true); err == nil {
		for _, item := range netStat {
			optionSet[item.Name] = struct{}{}
		}
	}
	if names, err := monitorRepo.GetNetworkNames(); err == nil {
		for _, name := range names {
			optionSet[name] = struct{}{}
		}
	}
	return sortedMonitorOptions(optionSet)
}

func (m *MonitorService) LoadSetting() (*dto.MonitorSetting, error) {
	setting, err := settingRepo.GetList()
	if err != nil {
		return nil, buserr.New("ErrRecordNotFound")
	}
	settingMap := make(map[string]string)
	for _, set := range setting {
		settingMap[set.Key] = set.Value
	}
	var info dto.MonitorSetting
	arr, err := json.Marshal(settingMap)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(arr, &info); err != nil {
		return nil, err
	}
	return &info, err
}

func (m *MonitorService) UpdateSetting(key, value string) error {
	monitorSettingMutex.Lock()
	defer monitorSettingMutex.Unlock()
	switch key {
	case "MonitorStatus":
		if value != constant.StatusEnable && value != constant.StatusDisable {
			return fmt.Errorf("invalid monitoring status")
		}
	case "MonitorInterval":
		interval, err := strconv.Atoi(value)
		if err != nil || interval < 10 || interval > 43200 {
			return fmt.Errorf("monitoring interval must be between 10 and 43200 seconds")
		}
	case "MonitorStoreDays":
		days, err := strconv.Atoi(value)
		if err != nil || days < 1 {
			return fmt.Errorf("monitoring retention must be a positive integer")
		}
	}
	switch key {
	case "MonitorStatus":
		if value == constant.StatusEnable && monitorCancel == nil {
			interval, err := settingRepo.Get(settingRepo.WithByKey("MonitorInterval"))
			if err != nil {
				return err
			}
			if err := StartMonitor(false, interval.Value); err != nil {
				return err
			}
		}
		if value == constant.StatusDisable && monitorCancel != nil {
			monitorCancel()
			monitorCancel = nil
			global.Cron.Remove(cron.EntryID(global.MonitorCronID))
			global.MonitorCronID = 0
		}
	case "MonitorInterval":
		status, err := settingRepo.Get(settingRepo.WithByKey("MonitorStatus"))
		if err != nil {
			return err
		}
		if status.Value == constant.StatusEnable && monitorCancel != nil {
			if err := StartMonitor(true, value); err != nil {
				return err
			}
		}
	}
	return settingRepo.Update(key, value)
}

func (m *MonitorService) CleanData(monitorType string) error {
	switch monitorType {
	case "host":
		return monitorRepo.CleanHost()
	default:
		return fmt.Errorf("unsupported monitoring cleanup type: %s", monitorType)
	}
}

func (m *MonitorService) Run() {
	if m.ctx.Err() != nil {
		return
	}
	var itemModel model.MonitorBase
	totalPercent, _ := cpu.Percent(3*time.Second, false)
	if len(totalPercent) == 1 {
		itemModel.Cpu = totalPercent[0]
	}
	topCPU := loadTopCPU()
	if len(topCPU) != 0 {
		topItemCPU, err := json.Marshal(topCPU)
		if err == nil {
			itemModel.TopCPU = string(topItemCPU)
		}
	}
	cpuCount, _ := psutil.CPUInfo.GetPhysicalCores(false)
	loadInfo, _ := load.Avg()
	itemModel.CpuLoad1 = loadInfo.Load1
	itemModel.CpuLoad5 = loadInfo.Load5
	itemModel.CpuLoad15 = loadInfo.Load15
	itemModel.LoadUsage = loadInfo.Load1 / (float64(cpuCount*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	itemModel.Memory = memoryInfo.UsedPercent
	topMem := loadTopMem()
	if len(topMem) != 0 {
		topMemItem, err := json.Marshal(topMem)
		if err == nil {
			itemModel.TopMem = string(topMemItem)
		}
	}

	if err := monitorRepo.CreateMonitorBase(itemModel); err != nil {
		global.LOG.Errorf("Insert basic monitoring data failed, err: %v", err)
	}

	m.loadDiskIO()
	m.loadNetIO()

	MonitorStoreDays, err := settingRepo.Get(settingRepo.WithByKey("MonitorStoreDays"))
	if err != nil {
		return
	}
	storeDays, _ := strconv.Atoi(MonitorStoreDays.Value)
	timeForDelete := time.Now().AddDate(0, 0, -storeDays)
	_ = monitorRepo.DelMonitorBase(timeForDelete)
	_ = monitorRepo.DelMonitorIO(timeForDelete)
	_ = monitorRepo.DelMonitorNet(timeForDelete)
}

func (m *MonitorService) loadDiskIO() {
	ioStat, _ := disk.IOCountersWithContext(m.ctx)
	var diskIOList []disk.IOCountersStat
	for _, io := range ioStat {
		diskIOList = append(diskIOList, io)
	}
	diskIOList = append(diskIOList, sumDiskIOCounters(ioStat))
	select {
	case <-m.ctx.Done():
	case m.DiskIO <- diskIOList:
	}
}

func (m *MonitorService) loadNetIO() {
	netStat, _ := net.IOCountersWithContext(m.ctx, true)
	netStatAll, _ := net.IOCountersWithContext(m.ctx, false)
	var netList []net.IOCountersStat
	netList = append(netList, netStat...)
	netList = append(netList, netStatAll...)
	select {
	case <-m.ctx.Done():
	case m.NetIO <- netList:
	}
}

func (m *MonitorService) saveIODataToDB(ctx context.Context, interval float64) {
	for {
		select {
		case <-ctx.Done():
			return
		case ioStat := <-m.DiskIO:
			select {
			case <-ctx.Done():
				return
			case ioStat2 := <-m.DiskIO:
				var ioList []model.MonitorIO
				for _, io2 := range ioStat2 {
					for _, io1 := range ioStat {
						if io2.Name == io1.Name {
							var itemIO model.MonitorIO
							itemIO.Name = io1.Name
							if io2.ReadBytes != 0 && io1.ReadBytes != 0 && io2.ReadBytes > io1.ReadBytes {
								itemIO.Read = uint64(float64(io2.ReadBytes-io1.ReadBytes) / interval)
							}
							if io2.WriteBytes != 0 && io1.WriteBytes != 0 && io2.WriteBytes > io1.WriteBytes {
								itemIO.Write = uint64(float64(io2.WriteBytes-io1.WriteBytes) / interval)
							}

							if io2.ReadCount != 0 && io1.ReadCount != 0 && io2.ReadCount > io1.ReadCount {
								itemIO.Count = uint64(float64(io2.ReadCount-io1.ReadCount) / interval)
							}
							writeCount := uint64(0)
							if io2.WriteCount != 0 && io1.WriteCount != 0 && io2.WriteCount > io1.WriteCount {
								writeCount = uint64(float64(io2.WriteCount-io1.WriteCount) / interval)
							}
							if writeCount > itemIO.Count {
								itemIO.Count = writeCount
							}

							if io2.ReadTime != 0 && io1.ReadTime != 0 && io2.ReadTime > io1.ReadTime {
								itemIO.Time = uint64(float64(io2.ReadTime-io1.ReadTime) / interval)
							}
							writeTime := uint64(0)
							if io2.WriteTime != 0 && io1.WriteTime != 0 && io2.WriteTime > io1.WriteTime {
								writeTime = uint64(float64(io2.WriteTime-io1.WriteTime) / interval)
							}
							if writeTime > itemIO.Time {
								itemIO.Time = writeTime
							}
							ioList = append(ioList, itemIO)
							break
						}
					}
				}
				_ = monitorRepo.BatchCreateMonitorIO(ioList)
				select {
				case <-ctx.Done():
					return
				case m.DiskIO <- ioStat2:
				}
			}
		}
	}
}

func (m *MonitorService) saveNetDataToDB(ctx context.Context, interval float64) {
	for {
		select {
		case <-ctx.Done():
			return
		case netStat := <-m.NetIO:
			select {
			case <-ctx.Done():
				return
			case netStat2 := <-m.NetIO:
				var netList []model.MonitorNetwork
				for _, net2 := range netStat2 {
					for _, net1 := range netStat {
						if net2.Name == net1.Name {
							var itemNet model.MonitorNetwork
							itemNet.Name = net1.Name

							if net2.BytesSent != 0 && net1.BytesSent != 0 && net2.BytesSent > net1.BytesSent {
								itemNet.Up = float64(net2.BytesSent-net1.BytesSent) / 1024 / interval
							}
							if net2.BytesRecv != 0 && net1.BytesRecv != 0 && net2.BytesRecv > net1.BytesRecv {
								itemNet.Down = float64(net2.BytesRecv-net1.BytesRecv) / 1024 / interval
							}
							netList = append(netList, itemNet)
							break
						}
					}
				}

				_ = monitorRepo.BatchCreateMonitorNet(netList)
				select {
				case <-ctx.Done():
					return
				case m.NetIO <- netStat2:
				}
			}
		}
	}
}

func loadTopCPU() []dto.Process {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	top5 := make([]dto.Process, 0, 5)
	for _, p := range processes {
		percent, err := p.CPUPercent()
		if err != nil {
			continue
		}
		minIndex := 0
		if len(top5) >= 5 {
			minCPU := top5[0].Percent
			for i := 1; i < len(top5); i++ {
				if top5[i].Percent < minCPU {
					minCPU = top5[i].Percent
					minIndex = i
				}
			}
			if percent < minCPU {
				continue
			}
		}
		name, err := p.Name()
		if err != nil {
			name = "undefined"
		}
		cmd, err := p.Cmdline()
		if err != nil {
			cmd = "undefined"
		}
		user, err := p.Username()
		if err != nil {
			user = "undefined"
		}
		if len(top5) == 5 {
			top5[minIndex] = dto.Process{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmd}
		} else {
			top5 = append(top5, dto.Process{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmd})
		}
	}
	sort.Slice(top5, func(i, j int) bool {
		return top5[i].Percent > top5[j].Percent
	})

	return top5
}

func loadTopMem() []dto.Process {
	processes, err := process.Processes()
	if err != nil {
		return nil
	}

	top5 := make([]dto.Process, 0, 5)
	for _, p := range processes {
		stat, err := p.MemoryInfo()
		if err != nil {
			continue
		}
		memItem := stat.RSS
		minIndex := 0
		if len(top5) >= 5 {
			min := top5[0].Memory
			for i := 1; i < len(top5); i++ {
				if top5[i].Memory < min {
					min = top5[i].Memory
					minIndex = i
				}
			}
			if memItem < min {
				continue
			}
		}
		name, err := p.Name()
		if err != nil {
			name = "undefined"
		}
		cmd, err := p.Cmdline()
		if err != nil {
			cmd = "undefined"
		}
		user, err := p.Username()
		if err != nil {
			user = "undefined"
		}
		percent, _ := p.MemoryPercent()
		if len(top5) == 5 {
			top5[minIndex] = dto.Process{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmd, Memory: memItem}
		} else {
			top5 = append(top5, dto.Process{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmd, Memory: memItem})
		}
	}

	sort.Slice(top5, func(i, j int) bool {
		return top5[i].Memory > top5[j].Memory
	})
	return top5
}

func StartMonitor(removeBefore bool, interval string) error {
	intervalItem, err := strconv.Atoi(interval)
	if err != nil || intervalItem < 10 || intervalItem > 43200 {
		return fmt.Errorf("invalid host monitoring interval: %s", interval)
	}
	if removeBefore {
		monitorCancel()
		global.Cron.Remove(cron.EntryID(global.MonitorCronID))
	}
	ctx, cancel := context.WithCancel(context.Background())
	service := &MonitorService{
		DiskIO: make(chan []disk.IOCountersStat, 2),
		NetIO:  make(chan []net.IOCountersStat, 2),
		ctx:    ctx,
	}
	now := time.Now()
	nextMinute := now.Truncate(time.Minute).Add(time.Minute)
	timer := time.AfterFunc(time.Until(nextMinute), func() {
		monitorSettingMutex.Lock()
		defer monitorSettingMutex.Unlock()
		if ctx.Err() != nil {
			return
		}
		monitorID, err := global.Cron.AddJob(fmt.Sprintf("@every %ss", interval), service)
		if err != nil {
			return
		}
		global.MonitorCronID = monitorID
	})
	monitorCancel = func() {
		cancel()
		timer.Stop()
	}

	service.Run()

	go service.saveIODataToDB(ctx, float64(intervalItem))
	go service.saveNetDataToDB(ctx, float64(intervalItem))

	return nil
}

func sumDiskIOCounters(ioStats map[string]disk.IOCountersStat) disk.IOCountersStat {
	total := disk.IOCountersStat{Name: "all"}
	for name, stat := range ioStats {
		if isBlockDevicePartition(name) {
			continue
		}

		total.ReadCount += stat.ReadCount
		total.MergedReadCount += stat.MergedReadCount
		total.WriteCount += stat.WriteCount
		total.MergedWriteCount += stat.MergedWriteCount
		total.ReadBytes += stat.ReadBytes
		total.WriteBytes += stat.WriteBytes
		total.ReadTime += stat.ReadTime
		total.WriteTime += stat.WriteTime
		total.IopsInProgress += stat.IopsInProgress
		total.IoTime += stat.IoTime
		total.WeightedIO += stat.WeightedIO
	}
	return total
}

func isBlockDevicePartition(name string) bool {
	deviceName := filepath.Base(name)
	if cached, ok := blockDevicePartitionCache.Load(deviceName); ok {
		return cached.(bool)
	}

	_, err := os.Stat(filepath.Join(hostSysPath, "class", "block", deviceName, "partition"))
	isPartition := err == nil
	actual, _ := blockDevicePartitionCache.LoadOrStore(deviceName, isPartition)
	return actual.(bool)
}

func loadHostSysPath() string {
	hostSys := os.Getenv("HOST_SYS")
	if hostSys == "" {
		return "/sys"
	}
	return hostSys
}

func sortedMonitorOptions(optionSet map[string]struct{}) []string {
	options := make([]string, 0, len(optionSet))
	for name := range optionSet {
		if len(name) != 0 && name != "all" {
			options = append(options, name)
		}
	}
	sort.Strings(options)
	return append([]string{"all"}, options...)
}
