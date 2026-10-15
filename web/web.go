// Package web: 前端产物嵌入。go:embed 无法跨目录, 因此 embed 声明放在本包。
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// Dist 返回 dist 子文件系统
func Dist() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
