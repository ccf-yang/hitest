package disks

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"hitest/internal/pkg/buserr"
	
	"hitest/internal/pkg/cmdx"

	
	
)

type DiskService struct{}

type IDiskService interface {
	GetCompleteDiskInfo() (*CompleteDiskInfo, error)
	PartitionDisk(req DiskPartitionRequest) (string, error)
	MountDisk(req DiskMountRequest) error
	UnmountDisk(req DiskUnmountRequest) error
}

func NewIDiskService() IDiskService {
	return &DiskService{}
}

func (s *DiskService) GetCompleteDiskInfo() (*CompleteDiskInfo, error) {
	var diskInfos []DiskBasicInfo
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Second))
	output, err := cmdMgr.RunWithStdout("lsblk", "-J", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL,SERIAL,TRAN,ROTA")
	if err == nil {
		diskInfos, err = parseLsblkJsonOutput(output)
		if err == nil {
			result := organizeDiskInfo(diskInfos)
			return &result, nil
		}
	}
	output, err = cmdMgr.RunWithStdout("lsblk", "-P", "-o", "NAME,SIZE,TYPE,MOUNTPOINT,FSTYPE,MODEL,SERIAL,TRAN,ROTA")
	if err != nil {
		return nil, fmt.Errorf("failed to run lsblk command: %v", err)
	}
	diskInfos, err = parseLsblkOutput(output)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lsblk output: %v", err)
	}
	result := organizeDiskInfo(diskInfos)
	return &result, nil
}

func (s *DiskService) PartitionDisk(req DiskPartitionRequest) (string, error) {
	if cmd.CheckIllegal(req.Device, req.Filesystem, req.MountPoint, req.Label) {
		return "", buserr.New("ErrCmdIllegal")
	}
	if !strings.HasPrefix("/dev", req.Device) {
		req.Device = "/dev/" + req.Device
	}
	if !deviceExists(req.Device) {
		return "", buserr.WithName("DeviceNotFound", req.Device)
	}

	if isDeviceMounted(req.Device) {
		return "", buserr.WithName("DeviceIsMounted", req.Device)
	}

	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
	if err := cmdMgr.Run("partprobe", req.Device); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}

	if err := cmdMgr.Run("parted", "-s", req.Device, "mklabel", "gpt"); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}

	if err := cmdMgr.Run("parted", "-s", req.Device, "mkpart", "primary", "1MiB", "100%"); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}

	if err := cmdMgr.Run("partprobe", req.Device); err != nil {
		return "", buserr.WithErr("PartitionDiskErr", err)
	}
	partition := req.Device + "1"

	formatReq := DiskFormatRequest{
		Device:     partition,
		Filesystem: req.Filesystem,
		Label:      req.Label,
	}
	if err := formatDisk(formatReq); err != nil {
		return "", buserr.WithErr("FormatDiskErr", err)
	}

	if req.MountPoint != "" {
		mountReq := DiskMountRequest{
			Device:     partition,
			MountPoint: req.MountPoint,
			Filesystem: req.Filesystem,
			AutoMount:  req.AutoMount,
		}
		if err := s.MountDisk(mountReq); err != nil {
			return "", buserr.WithErr("MountDiskErr", err)
		}
	}

	return partition, nil
}

func (s *DiskService) MountDisk(req DiskMountRequest) error {
	if cmd.CheckIllegal(req.Device, req.MountPoint, req.Filesystem) {
		return buserr.New("ErrCmdIllegal")
	}
	if !deviceExists(req.Device) {
		return buserr.WithName("DeviceNotFound", req.Device)
	}

	if err := os.MkdirAll(req.MountPoint, 0755); err != nil {
		return err
	}

	fileSystem, err := getFilesystemType(req.Device)
	if err != nil {
		return err
	}
	if fileSystem == "" {
		formatReq := DiskFormatRequest{
			Device:     req.Device,
			Filesystem: req.Filesystem,
		}
		if err := formatDisk(formatReq); err != nil {
			return buserr.WithErr("FormatDiskErr", err)
		}
	}

	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(1 * time.Minute))
	if err := cmdMgr.Run("mount", "-t", req.Filesystem, req.Device, req.MountPoint); err != nil {
		return buserr.WithErr("MountDiskErr", err)
	}
	if req.AutoMount {
		options := ""
		if req.NoFail {
			options = "defaults,nofail"
		}
		if err := addToFstabWithOptions(req.Device, req.MountPoint, req.Filesystem, options); err != nil {
			return buserr.WithErr("MountDiskErr", err)
		}
	}

	return nil
}

func (s *DiskService) UnmountDisk(req DiskUnmountRequest) error {
	if cmd.CheckIllegal(req.MountPoint) {
		return buserr.New("ErrCmdIllegal")
	}
	if !isPointMounted(req.MountPoint) {
		return buserr.New("MountDiskErr")
	}
	if err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).Run("umount", "-f", req.MountPoint); err != nil {
		return buserr.WithErr("MountDiskErr", err)
	}
	if err := removeFromFstab(req.MountPoint); err != nil {
		log.Printf("remove %s mountPoint err: %v", req.MountPoint, err)
	}
	return nil
}
