package psutil

import (
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/disk"
)

const (
	diskUsageCacheInterval     = 30 * time.Second
	diskPartitionCacheInterval = 10 * time.Minute
)

type DiskUsageEntry struct {
	lastSampleTime time.Time
	cachedUsage    *disk.UsageStat
}

type DiskState struct {
	usageMu    sync.RWMutex
	usageCache map[string]*DiskUsageEntry

	partitionMu       sync.RWMutex
	lastPartitionTime time.Time
	cachedPartitions  []disk.PartitionStat
}

func (d *DiskState) GetUsage(path string, forceRefresh bool) (*disk.UsageStat, error) {
	d.usageMu.RLock()
	if entry, ok := d.usageCache[path]; ok {
		if time.Since(entry.lastSampleTime) < diskUsageCacheInterval && !forceRefresh {
			defer d.usageMu.RUnlock()
			return entry.cachedUsage, nil
		}
	}
	d.usageMu.RUnlock()

	usage, err := disk.Usage(path)
	if err != nil {
		return nil, err
	}

	d.usageMu.Lock()
	if d.usageCache == nil {
		d.usageCache = make(map[string]*DiskUsageEntry)
	}
	d.usageCache[path] = &DiskUsageEntry{
		lastSampleTime: time.Now(),
		cachedUsage:    usage,
	}
	d.usageMu.Unlock()

	return usage, nil
}
