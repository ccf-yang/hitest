package diskinfo

import (
	"fmt"
	"log"
	network "net"
	"sort"
	"strings"
	"sync"
	"time"

	"hitest/internal/modules/dto"
	"hitest/internal/pkg/cmdx"
	"hitest/internal/pkg/psutil"

	"github.com/shirou/gopsutil/v4/disk"
)

type diskInfo struct {
	Type   string
	Mount  string
	Device string
}

func DiskInfo(forceRefresh bool) []dto.DiskInfo {
	var datas []dto.DiskInfo
	stdout := Mounts()
	lines := strings.Split(stdout, "\n")

	var mounts []diskInfo
	var excludes = []string{"/mnt/cdrom", "/boot", "/boot/efi", "/dev", "/dev/shm", "/run/lock", "/run", "/run/shm", "/run/user"}
	for _, line := range lines {
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		if strings.HasPrefix(fields[6], "/snap") || len(strings.Split(fields[6], "/")) > 10 {
			continue
		}
		if strings.TrimSpace(fields[1]) == "tmpfs" || strings.TrimSpace(fields[1]) == "overlay" {
			continue
		}
		if strings.Contains(fields[2], "K") {
			continue
		}
		if strings.Contains(fields[6], "docker") || strings.Contains(fields[6], "podman") || strings.Contains(fields[6], "containerd") || strings.HasPrefix(fields[6], "/var/lib/containers") {
			continue
		}
		isExclude := false
		for _, exclude := range excludes {
			if exclude == fields[6] {
				isExclude = true
			}
		}
		if isExclude {
			continue
		}
		mounts = append(mounts, diskInfo{Type: fields[1], Device: fields[0], Mount: strings.Join(fields[6:], " ")})
	}

	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)
	wg.Add(len(mounts))
	for i := 0; i < len(mounts); i++ {
		go func(mount diskInfo) {
			defer wg.Done()

			var itemData dto.DiskInfo
			itemData.Path = mount.Mount
			itemData.Type = mount.Type
			itemData.Device = mount.Device

			state, err := UsageWithTimeout(mount.Mount, forceRefresh)
			if err != nil {
				log.Printf("load disk info from %s failed, err: %v", mount.Mount, err)
			} else {
				itemData.Total = state.Total
				itemData.Free = state.Free
				itemData.Used = state.Used
				itemData.UsedPercent = state.UsedPercent
				itemData.InodesTotal = state.InodesTotal
				itemData.InodesUsed = state.InodesUsed
				itemData.InodesFree = state.InodesFree
				itemData.InodesUsedPercent = state.InodesUsedPercent
			}
			mu.Lock()
			datas = append(datas, itemData)
			mu.Unlock()
		}(mounts[i])
	}
	wg.Wait()

	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Path < datas[j].Path
	})
	return datas
}

var diskMountsMu sync.Mutex

func Mounts() string {
	if !diskMountsMu.TryLock() {
		return ""
	}
	resultCh := make(chan string, 1)
	go func() {
		var stdout string
		defer func() {
			diskMountsMu.Unlock()
			resultCh <- stdout
		}()
		cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(2 * time.Second))
		format := `NR>1 && !/tmpfs|snap\/core|udev/ {printf "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", $1, $2, $3, $4, $5, $6, $7}`
		output, err := cmdMgr.RunPipe(
			cmd.PipeCommand{Name: "df", Args: []string{"-hT", "-P"}},
			cmd.PipeCommand{Name: "awk", Args: []string{format}},
		)
		if err != nil {
			log.Printf("load disk info with df -hT -P failed, err: %v", err)
			cmdMgr2 := cmd.NewCommandMgr(cmd.WithTimeout(1 * time.Second))
			output, err = cmdMgr2.RunPipe(
				cmd.PipeCommand{Name: "df", Args: []string{"-lhT", "-P"}},
				cmd.PipeCommand{Name: "awk", Args: []string{format}},
			)
			if err != nil {
				log.Printf("load disk info with df -lhT -P failed, err: %v", err)
				return
			}
		}

		stdout = output
	}()
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case stdout := <-resultCh:
		return stdout
	case <-timer.C:
		log.Println("load disk mounts timed out; df collection is still running")
		return ""
	}
}

func UsageWithTimeout(path string, forceRefresh bool) (*disk.UsageStat, error) {
	type diskResult struct {
		state *disk.UsageStat
		err   error
	}
	resultCh := make(chan diskResult, 1)
	go func() {
		state, err := psutil.DISK.GetUsage(path, forceRefresh)
		resultCh <- diskResult{state: state, err: err}
	}()
	select {
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("load disk usage from %s: timeout", path)
	case result := <-resultCh:
		return result.state, result.err
	}
}

func loadOutboundIP() string {
	conn, err := network.Dial("udp", "8.8.8.8:80")

	if err != nil {
		return "IPNotFound"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*network.UDPAddr)
	return localAddr.IP.String()
}
