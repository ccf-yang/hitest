package monitor

// 采样器: 移植自 legacy service/monitor.go。
// 变更: repo 调用内联为 gorm 直查(挂在模块的 *store.Store 上),
// global.Cron 换成 ticker goroutine, 设置存 store KV。

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"hitest/internal/modules/dto"
	"hitest/internal/pkg/psutil"
	"hitest/internal/store"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/load"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	defaultInterval  = "300"
	defaultStoreDays = "30"
)

type sampler struct {
	st     *store.Store
	DiskIO chan []disk.IOCountersStat
	NetIO  chan []net.IOCountersStat
	ctx    context.Context
}

var (
	settingMutex   sync.Mutex
	monitorCancel  context.CancelFunc
	hostSysPath    = loadHostSysPath()
	partitionCache sync.Map
)

// ensureDefaults 首次调用时写入默认设置
func ensureDefaults(st *store.Store) {
	if st.KVGet("MonitorInterval") == "" {
		st.KVSet("MonitorInterval", defaultInterval)
	}
	if st.KVGet("MonitorStoreDays") == "" {
		st.KVSet("MonitorStoreDays", defaultStoreDays)
	}
	if st.KVGet("MonitorStatus") == "" {
		st.KVSet("MonitorStatus", "Enable")
	}
}

// StartSampler 启动历史采样（模块注册时调用; 幂等）
func StartSampler(st *store.Store) {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	ensureDefaults(st)
	if monitorCancel != nil {
		return // 已在运行
	}
	if st.KVGet("MonitorStatus") != "Enable" {
		return
	}
	interval := st.KVGet("MonitorInterval")
	if err := startMonitor(st, interval); err != nil {
		log.Printf("[monitor] 采样启动失败: %v", err)
	}
}

func startMonitor(st *store.Store, intervalStr string) error {
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 10 || interval > 43200 {
		return fmt.Errorf("invalid monitoring interval: %s", intervalStr)
	}
	ctx, cancel := context.WithCancel(context.Background())
	s := &sampler{
		st:     st,
		DiskIO: make(chan []disk.IOCountersStat, 2),
		NetIO:  make(chan []net.IOCountersStat, 2),
		ctx:    ctx,
	}
	monitorCancel = cancel

	iv := time.Duration(interval) * time.Second
	go func() {
		s.Run() // 立即采一次
		ticker := time.NewTicker(iv)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.Run()
			}
		}
	}()
	go s.saveIODataToDB(float64(interval))
	go s.saveNetDataToDB(float64(interval))
	return nil
}

func stopMonitor() {
	settingMutex.Lock()
	defer settingMutex.Unlock()
	if monitorCancel != nil {
		monitorCancel()
		monitorCancel = nil
	}
}

// ---------- 采样一次 (legacy Run) ----------

func (s *sampler) Run() {
	if s.ctx.Err() != nil {
		return
	}
	var item MonitorBase
	totalPercent, _ := cpu.Percent(3*time.Second, false)
	if len(totalPercent) == 1 {
		item.Cpu = totalPercent[0]
	}
	topCPU := loadTopCPU()
	if len(topCPU) != 0 {
		if b, err := json.Marshal(topCPU); err == nil {
			item.TopCPU = string(b)
		}
	}
	cpuCount, _ := psutil.CPUInfo.GetPhysicalCores(false)
	loadInfo, _ := load.Avg()
	item.CpuLoad1 = loadInfo.Load1
	item.CpuLoad5 = loadInfo.Load5
	item.CpuLoad15 = loadInfo.Load15
	item.LoadUsage = loadInfo.Load1 / (float64(cpuCount*2) * 0.75) * 100

	memoryInfo, _ := mem.VirtualMemory()
	item.Memory = memoryInfo.UsedPercent
	topMem := loadTopMem()
	if len(topMem) != 0 {
		if b, err := json.Marshal(topMem); err == nil {
			item.TopMem = string(b)
		}
	}

	if err := s.st.DB().Create(&item).Error; err != nil {
		log.Printf("[monitor] 写入基础监控数据失败: %v", err)
	}

	s.loadDiskIO()
	s.loadNetIO()

	storeDays, _ := strconv.Atoi(s.st.KVGet("MonitorStoreDays"))
	if storeDays <= 0 {
		storeDays = 30
	}
	cutoff := time.Now().AddDate(0, 0, -storeDays)
	_ = s.st.DB().Where("created_at < ?", cutoff).Delete(&MonitorBase{}).Error
	_ = s.st.DB().Where("created_at < ?", cutoff).Delete(&MonitorIO{}).Error
	_ = s.st.DB().Where("created_at < ?", cutoff).Delete(&MonitorNetwork{}).Error
}

func (s *sampler) loadDiskIO() {
	ioStat, _ := disk.IOCountersWithContext(s.ctx)
	var diskIOList []disk.IOCountersStat
	for _, io := range ioStat {
		diskIOList = append(diskIOList, io)
	}
	diskIOList = append(diskIOList, sumDiskIOCounters(ioStat))
	select {
	case <-s.ctx.Done():
	case s.DiskIO <- diskIOList:
	}
}

func (s *sampler) loadNetIO() {
	netStat, _ := net.IOCountersWithContext(s.ctx, true)
	netStatAll, _ := net.IOCountersWithContext(s.ctx, false)
	var netList []net.IOCountersStat
	netList = append(netList, netStat...)
	netList = append(netList, netStatAll...)
	select {
	case <-s.ctx.Done():
	case s.NetIO <- netList:
	}
}

func (s *sampler) saveIODataToDB(interval float64) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case ioStat := <-s.DiskIO:
			select {
			case <-s.ctx.Done():
				return
			case ioStat2 := <-s.DiskIO:
				var ioList []MonitorIO
				for _, io2 := range ioStat2 {
					for _, io1 := range ioStat {
						if io2.Name != io1.Name {
							continue
						}
						var itemIO MonitorIO
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
				if len(ioList) > 0 {
					_ = s.st.DB().Create(&ioList).Error
				}
				select {
				case <-s.ctx.Done():
					return
				case s.DiskIO <- ioStat2:
				}
			}
		}
	}
}

func (s *sampler) saveNetDataToDB(interval float64) {
	for {
		select {
		case <-s.ctx.Done():
			return
		case netStat := <-s.NetIO:
			select {
			case <-s.ctx.Done():
				return
			case netStat2 := <-s.NetIO:
				var netList []MonitorNetwork
				for _, net2 := range netStat2 {
					for _, net1 := range netStat {
						if net2.Name != net1.Name {
							continue
						}
						var itemNet MonitorNetwork
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
				if len(netList) > 0 {
					_ = s.st.DB().Create(&netList).Error
				}
				select {
				case <-s.ctx.Done():
					return
				case s.NetIO <- netStat2:
				}
			}
		}
	}
}

// ---------- Top 进程 (legacy loadTopCPU/loadTopMem) ----------

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
		cmdline, err := p.Cmdline()
		if err != nil {
			cmdline = "undefined"
		}
		user, err := p.Username()
		if err != nil {
			user = "undefined"
		}
		item := dto.Process{Percent: percent, Pid: p.Pid, User: user, Name: name, Cmd: cmdline}
		if len(top5) == 5 {
			top5[minIndex] = item
		} else {
			top5 = append(top5, item)
		}
	}
	sort.Slice(top5, func(i, j int) bool { return top5[i].Percent > top5[j].Percent })
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
		cmdline, err := p.Cmdline()
		if err != nil {
			cmdline = "undefined"
		}
		user, err := p.Username()
		if err != nil {
			user = "undefined"
		}
		percent, _ := p.MemoryPercent()
		item := dto.Process{Percent: float64(percent), Pid: p.Pid, User: user, Name: name, Cmd: cmdline, Memory: memItem}
		if len(top5) == 5 {
			top5[minIndex] = item
		} else {
			top5 = append(top5, item)
		}
	}
	sort.Slice(top5, func(i, j int) bool { return top5[i].Memory > top5[j].Memory })
	return top5
}

// ---------- 磁盘 IO 汇总 (legacy) ----------

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
	if cached, ok := partitionCache.Load(deviceName); ok {
		return cached.(bool)
	}
	_, err := os.Stat(filepath.Join(hostSysPath, "class", "block", deviceName, "partition"))
	isPartition := err == nil
	actual, _ := partitionCache.LoadOrStore(deviceName, isPartition)
	return actual.(bool)
}

func loadHostSysPath() string {
	hostSys := os.Getenv("HOST_SYS")
	if hostSys == "" {
		return "/sys"
	}
	return hostSys
}

// ---------- 选项辅助 ----------

func diskIOCounters() ([]string, error) {
	stat, err := disk.IOCounters()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(stat))
	for name := range stat {
		names = append(names, name)
	}
	return names, nil
}

func netIOCounters() []string {
	stat, _ := net.IOCounters(true)
	names := make([]string, 0, len(stat))
	for _, s := range stat {
		names = append(names, s.Name)
	}
	return names
}

func sortedOptions(optionSet map[string]struct{}) []string {
	options := make([]string, 0, len(optionSet))
	for name := range optionSet {
		if len(name) != 0 && name != "all" {
			options = append(options, name)
		}
	}
	sort.Strings(options)
	return append([]string{"all"}, options...)
}
