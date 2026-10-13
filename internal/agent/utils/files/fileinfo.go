package files

import (
	"io/fs"
	"os"
	"time"

	"github.com/spf13/afero"
)

var binaryPreviewMimeTypes = map[string]struct{}{
	"application/pdf":               {},
	"application/zip":               {},
	"application/gzip":              {},
	"application/x-gzip":            {},
	"application/x-7z-compressed":   {},
	"application/x-rar-compressed":  {},
	"application/x-bzip2":           {},
	"application/x-xz":              {},
	"application/x-tar":             {},
	"application/java-archive":      {},
	"application/msword":            {},
	"application/vnd.ms-excel":      {},
	"application/vnd.ms-powerpoint": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
}

var binaryPreviewMimePrefixes = []string{
	"image/",
	"audio/",
	"video/",
	"application/vnd.ms-",
	"application/vnd.openxmlformats-officedocument.",
	"application/vnd.oasis.opendocument.",
}

var binaryPreviewExtensions = map[string]struct{}{
	".pdf":  {},
	".zip":  {},
	".gz":   {},
	".bz2":  {},
	".xz":   {},
	".tar":  {},
	".tgz":  {},
	".rar":  {},
	".7z":   {},
	".war":  {},
	".doc":  {},
	".docx": {},
	".xls":  {},
	".xlsx": {},
	".ppt":  {},
	".pptx": {},
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".bmp":  {},
	".gif":  {},
	".tiff": {},
	".ico":  {},
	".webp": {},
	".svg":  {},
	".mp3":  {},
	".wav":  {},
	".wma":  {},
	".ape":  {},
	".acc":  {},
	".ogg":  {},
	".flac": {},
	".mp4":  {},
	".webm": {},
	".mov":  {},
	".wmv":  {},
	".mkv":  {},
	".avi":  {},
	".flv":  {},
}

type FileInfo struct {
	Fs           afero.Fs    `json:"-"`
	Path         string      `json:"path"`
	Name         string      `json:"name"`
	User         string      `json:"user"`
	Group        string      `json:"group"`
	Uid          string      `json:"uid"`
	Gid          string      `json:"gid"`
	Extension    string      `json:"extension"`
	Content      string      `json:"content"`
	Size         int64       `json:"size"`
	IsDir        bool        `json:"isDir"`
	IsSymlink    bool        `json:"isSymlink"`
	IsHidden     bool        `json:"isHidden"`
	LinkPath     string      `json:"linkPath"`
	Type         string      `json:"type"`
	Mode         string      `json:"mode"`
	MimeType     string      `json:"mimeType"`
	UpdateTime   time.Time   `json:"updateTime"`
	ModTime      time.Time   `json:"modTime"`
	FileMode     os.FileMode `json:"-"`
	Items        []*FileInfo `json:"items"`
	ItemTotal    int         `json:"itemTotal"`
	FavoriteID   uint        `json:"favoriteID"`
	ShareCode    string      `json:"shareCode"`
	IsDetail     bool        `json:"isDetail"`
	IsAppendOnly bool        `json:"isAppendOnly"`
	IsImmutable  bool        `json:"isImmutable"`
}

type FileOption struct {
	Path       string `json:"path"`
	Search     string `json:"search"`
	ContainSub bool   `json:"containSub"`
	Expand     bool   `json:"expand"`
	Dir        bool   `json:"dir"`
	ShowHidden bool   `json:"showHidden"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	SortBy     string `json:"sortBy"`
	SortOrder  string `json:"sortOrder"`
	IsDetail   bool   `json:"isDetail"`
}

type FileSearchInfo struct {
	Path string `json:"path"`
	fs.FileInfo
}

type CompressType string

const (
	Zip      CompressType = "zip"
	Gz       CompressType = "gz"
	Bz2      CompressType = "bz2"
	TarBz2   CompressType = "tar.bz2"
	Tar      CompressType = "tar"
	TarGz    CompressType = "tar.gz"
	Tgz      CompressType = "tgz"
	Xz       CompressType = "xz"
	TarXz    CompressType = "tar.xz"
	SdkZip   CompressType = "sdkZip"
	SdkTarGz CompressType = "sdkTarGz"
	Rar      CompressType = "rar"
	X7z      CompressType = "7z"
)
