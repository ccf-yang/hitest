package disks

// 移植自 legacy dto/request/disk.go + dto/response/disk.go

type DiskPartitionRequest struct {
	Device     string `json:"device" validate:"required"`
	Filesystem string `json:"filesystem" validate:"required,oneof=ext4 xfs"`
	Label      string `json:"label"`
	AutoMount  bool   `json:"autoMount"`
	MountPoint string `json:"mountPoint" validate:"required"`
}

type DiskMountRequest struct {
	Device     string `json:"device" validate:"required"`
	MountPoint string `json:"mountPoint" validate:"required"`
	Filesystem string `json:"filesystem" validate:"required,oneof=ext4 xfs"`
	AutoMount  bool   `json:"autoMount"`
	NoFail     bool   `json:"noFail"`
}

type DiskUnmountRequest struct {
	MountPoint string `json:"mountPoint" validate:"required"`
}

type DiskFormatRequest struct {
	Device     string `json:"device"`
	Label      string `json:"label"`
	MountPoint string `json:"mountPoint" validate:"required"`
	Filesystem string `json:"filesystem" validate:"required,oneof=ext4 xfs"`
}

type DiskBasicInfo struct {
	Device      string `json:"device"`
	Size        string `json:"size"`
	Model       string `json:"model"`
	DiskType    string `json:"diskType"`
	IsRemovable bool   `json:"isRemovable"`
	IsSystem    bool   `json:"isSystem"`
	Filesystem  string `json:"filesystem"`
	Used        string `json:"used"`
	Avail       string `json:"avail"`
	UsePercent  int    `json:"usePercent"`
	MountPoint  string `json:"mountPoint"`
	IsMounted   bool   `json:"isMounted"`
	Serial      string `json:"serial"`
}

type DiskInfo struct {
	DiskBasicInfo
	Partitions []DiskBasicInfo `json:"partitions"`
}

type CompleteDiskInfo struct {
	Disks              []DiskInfo      `json:"disks"`
	UnpartitionedDisks []DiskBasicInfo `json:"unpartitionedDisks"`
	SystemDisks        []DiskInfo      `json:"systemDisks"`
	TotalDisks         int             `json:"totalDisks"`
	TotalCapacity      int64           `json:"totalCapacity"`
}

type MountInfo struct {
	Device     string `json:"device"`
	MountPoint string `json:"mountPoint"`
	Filesystem string `json:"filesystem"`
	Options    string `json:"options"`
}
