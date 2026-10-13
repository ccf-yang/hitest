package files

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"hitest/internal/agent/utils/req_helper"

	"hitest/internal/agent/global"

	"github.com/spf13/afero"
	"golang.org/x/sync/singleflight"
)

const (
	cmdDefaultTimeout           = 10 * time.Second
	cmdRecursiveTimeout         = 5 * time.Minute
	maxArchiveSymlinkTargetSize = 4 * 1024
)

var protectedPaths = []string{
	"/",
	"/bin",
	"/sbin",
	"/etc",
	"/boot",
	"/usr",
	"/lib",
	"/lib64",
	"/dev",
	"/proc",
	"/sys",
	"/root",
}

var (
	dirSizeGroup   singleflight.Group
	dirSizeLimiter = make(chan struct{}, 2)
)

type FileOp struct {
	Fs afero.Fs
}

func NewFileOp() FileOp {
	return FileOp{
		Fs: afero.NewOsFs(),
	}
}

func (f FileOp) CreateDir(dst string, mode fs.FileMode) error {
	return f.Fs.MkdirAll(dst, mode)
}

func (f FileOp) CreateDirWithPath(isDir bool, pathItem string) (string, error) {
	checkPath := pathItem
	if !isDir {
		checkPath = path.Dir(pathItem)
	}
	if !f.Stat(checkPath) {
		if err := f.CreateDir(checkPath, os.ModePerm); err != nil {
			return pathItem, err
		}
	}
	return pathItem, nil
}

func (f FileOp) Stat(dst string) bool {
	info, _ := f.Fs.Stat(dst)
	return info != nil
}

type downloadTask struct {
	mu         sync.Mutex
	cancel     context.CancelFunc
	done       chan struct{}
	dst        string
	finished   bool
	cleanupErr error
}

var (
	downloadMu    sync.Mutex
	downloadTasks = make(map[string]*downloadTask)
)

type Process struct {
	Key     string  `json:"key,omitempty"`
	Total   uint64  `json:"total"`
	Written uint64  `json:"written"`
	Percent float64 `json:"percent"`
	Name    string  `json:"name"`
	Status  string  `json:"status,omitempty"`
	Attempt int     `json:"attempt,omitempty"`
	Error   string  `json:"error,omitempty"`
}

type DownloadProxyConfig struct {
	Type     string
	URL      string
	Port     string
	User     string
	Password string
}

type DownloadOptions struct {
	IgnoreCertificate bool
	UseServerFilename bool
	Proxy             *DownloadProxyConfig
}

type downloadState struct {
	written  int64
	total    int64
	etag     string
	finalURL string
}

type downloadPolicy struct {
	retries     int
	retryDelay  time.Duration
	idleTimeout time.Duration
}

var remoteDownloadPolicy = downloadPolicy{retries: 3, retryDelay: 5 * time.Second, idleTimeout: 90 * time.Second}

type downloadPartRecord struct {
	Path   string
	Device uint64
	Inode  uint64
}

func removeDownloadPart(part string, original os.FileInfo) error {
	current, err := os.Lstat(part)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if original == nil || !current.Mode().IsRegular() || !os.SameFile(original, current) {
		return fmt.Errorf("download temporary file changed: %s", part)
	}
	return os.Remove(part)
}

func validateDownloadRecordDir(dir string) error {
	info, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !info.IsDir() || info.Mode().Perm()&0077 != 0 || !ok || int(stat.Uid) != os.Geteuid() {
		return fmt.Errorf("download recovery directory must be private and owned by the agent")
	}
	return nil
}

// Run before accepting requests. Only recorded, identity-matched partial files may be removed.
func CleanupInterruptedDownloads() error {
	dir := filepath.Join(global.Dir.TmpDir, "remote-downloads")
	if err := validateDownloadRecordDir(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var errs []error
	for _, entry := range entries {
		if !entry.Type().IsRegular() || !strings.HasSuffix(entry.Name(), ".part.json") {
			continue
		}
		recordPath := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(recordPath)
		var record downloadPartRecord
		if err != nil || json.Unmarshal(data, &record) != nil || !filepath.IsAbs(record.Path) ||
			!strings.HasPrefix(filepath.Base(record.Path), ".hitest-download-") ||
			filepath.Base(record.Path)+".json" != entry.Name() {
			errs = append(errs, fmt.Errorf("invalid download recovery record: %s", entry.Name()))
			continue
		}
		info, err := os.Lstat(record.Path)
		if err == nil {
			stat, ok := info.Sys().(*syscall.Stat_t)
			if !ok || !info.Mode().IsRegular() || uint64(stat.Dev) != record.Device || uint64(stat.Ino) != record.Inode {
				errs = append(errs, fmt.Errorf("download recovery file identity changed: %s", record.Path))
				continue
			}
			err = removeDownloadPart(record.Path, info)
		}
		if err != nil && !os.IsNotExist(err) {
			errs = append(errs, err)
			continue
		}
		if err := os.Remove(recordPath); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (f FileOp) DownloadFile(url, dst string) error {
	resp, err := req_helper.HandleGet(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create download file [%s] error, err %s", dst, err.Error())
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("save download file [%s] error, err %s", dst, err.Error())
	}
	return nil
}

type DirSize struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
}

type contextAwareWriter struct {
	ctx    context.Context
	writer io.Writer
}

type DecompressOptions struct {
	PreserveOwner     bool
	AllowCLIReextract bool
}

type archiveOwnership struct {
	uid int
	gid int
}

const zipUnixOwnershipExtraID = 0x7875

type ownershipPreservingShellExtractor interface {
	ExtractWithOptions(ctx context.Context, filePath, dstDir, secret string, preserveOwner bool) error
}

type contextReader struct {
	ctx context.Context
	r   io.Reader
}
