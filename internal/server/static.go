package server

import (
	"io/fs"
	"net/http"
	"strings"

	"hitest/web"
)

// StaticHandler 托管内嵌的 Vue SPA, 未知路径回退 index.html。
func StaticHandler() http.Handler {
	sub, err := web.Dist()
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(sub, path); err != nil {
			if path != "index.html" {
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			http.Error(w, "前端未构建: 请先执行 make web-build", http.StatusNotFound)
			return
		}
		fileServer.ServeHTTP(w, r)
	})
}
