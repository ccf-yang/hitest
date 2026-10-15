// Package files: 文件管理。全部为流式 HTTP：
//   列目录 GET /files/list?path=          （JSON）
//   上传   PUT /files/file?path=          （请求体 = 文件内容, 流式落盘）
//   下载   GET /files/file?path=          （响应体 = 文件内容, 支持 Range）
//   建目录/改名/删除 POST /files/mkdir|rename|delete （JSON）
// 无自研分块协议——curl 直接可用, 大文件天然支持。
package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/gin-gonic/gin"

	"hitest/internal/store"
)

type FilesModule struct{}

func (m *FilesModule) Name() string { return "files" }

func (m *FilesModule) RegisterRoutes(api gin.IRouter, _ *store.Store) {
	api.GET("/files/list", m.list)
	api.PUT("/files/file", m.upload)
	api.GET("/files/file", m.download)
	api.POST("/files/mkdir", m.mkdir)
	api.POST("/files/rename", m.rename)
	api.POST("/files/delete", m.delete)
}

type Entry struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	Mode  string `json:"mode"`
	MTime int64  `json:"mtime"`
	IsDir bool   `json:"is_dir"`
}

func (m *FilesModule) list(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "/"
	}
	st, err := os.Stat(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !st.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "not a directory: " + path})
		return
	}
	dirents, err := os.ReadDir(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	entries := make([]Entry, 0, len(dirents))
	for _, de := range dirents {
		info, err := de.Info()
		if err != nil {
			continue // 列表期间被删除的文件跳过
		}
		entries = append(entries, Entry{
			Name:  de.Name(),
			Size:  info.Size(),
			Mode:  info.Mode().String(),
			MTime: info.ModTime().Unix(),
			IsDir: de.IsDir(),
		})
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir
		}
		return entries[i].Name < entries[j].Name
	})
	c.JSON(http.StatusOK, gin.H{"path": path, "entries": entries})
}

// upload: PUT 请求体即文件内容（流式, 不驻内存）
func (m *FilesModule) upload(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path"})
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	f, err := os.Create(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()
	if _, err := io.Copy(f, c.Request.Body); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true, "path": path})
}

// download: 响应体即文件内容（http.ServeContent 支持 Range 断点续传）
func (m *FilesModule) download(c *gin.Context) {
	path := c.Query("path")
	f, err := os.Open(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if st.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "is a directory: " + path})
		return
	}
	http.ServeContent(c.Writer, c.Request, st.Name(), st.ModTime(), f)
}

func (m *FilesModule) mkdir(c *gin.Context) {
	var req struct{ Path string `json:"path"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path"})
		return
	}
	if err := os.MkdirAll(req.Path, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *FilesModule) rename(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		NewPath string `json:"new_path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" || req.NewPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("bad request: %v", err)})
		return
	}
	if err := os.Rename(req.Path, req.NewPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (m *FilesModule) delete(c *gin.Context) {
	var req struct{ Path string `json:"path"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing path"})
		return
	}
	if err := os.RemoveAll(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
