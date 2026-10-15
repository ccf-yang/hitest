package client

import (
	"path"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// do: 带鉴权头的 HTTP 请求, 非 2xx 解析 {"error":...}
func (c *Config) do(method, url string, body io.Reader, contentType string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	if c.Token != "" {
		req.Header.Set("X-Token", c.Token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		var e struct {
			Error string `json:"error"`
		}
		if json.NewDecoder(resp.Body).Decode(&e) == nil && e.Error != "" {
			return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, e.Error)
		}
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return resp, nil
}

// Ls: GET /files/list
func (c *Config) Ls(machine, path string) error {
	if path == "" {
		path = "/"
	}
	resp, err := c.do("GET", c.machineURL(machine)+"/files/list?path="+path, nil, "")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Entries []struct {
			Name  string `json:"name"`
			Size  int64  `json:"size"`
			MTime int64  `json:"mtime"`
			IsDir bool   `json:"is_dir"`
		} `json:"entries"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	for _, e := range res.Entries {
		kind := "f"
		if e.IsDir {
			kind = "d"
		}
		fmt.Printf("%s %10d %s %s\n", kind, e.Size,
			time.Unix(e.MTime, 0).Format("2006-01-02 15:04"), e.Name)
	}
	return nil
}

// Exec: POST /exec/run
func (c *Config) Exec(machine, cmd string, timeoutSec int, asJSON bool) error {
	body, _ := json.Marshal(map[string]any{"cmd": cmd, "timeout_sec": timeoutSec})
	resp, err := c.do("POST", c.machineURL(machine)+"/exec/run", bytes.NewReader(body), "application/json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var res struct {
		Code   int    `json:"code"`
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		DurMs  int64  `json:"dur_ms"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	if asJSON {
		fmt.Println(prettyJSON(res))
	} else {
		os.Stdout.WriteString(res.Stdout)
		os.Stderr.WriteString(res.Stderr)
	}
	if res.Code != 0 {
		return fmt.Errorf("exit code %d", res.Code)
	}
	return nil
}

// Cp: 上传文件/目录到机器。dst 形如 "nas1:/opt/app/" 或 "nas1,nas2:/path"（多机并行）。
func (c *Config) Cp(src, dst string, recursive bool) error {
	machines, dstPath, err := parseDst(dst)
	if err != nil {
		return err
	}

	// 收集要传的文件（相对路径 → 绝对源路径）
	var items []item
	st, err := os.Stat(src)
	if err != nil {
		return err
	}
	if st.IsDir() {
		if !recursive {
			return fmt.Errorf("%s 是目录, 加 -r", src)
		}
		err = filepath.WalkDir(src, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(src, p)
			items = append(items, item{rel, p})
			return nil
		})
		if err != nil {
			return err
		}
	} else {
		items = []item{{filepath.Base(src), src}}
	}

	// 多机并行
	type result struct {
		machine string
		err     error
	}
	errCh := make(chan result, len(machines))
	for _, m := range machines {
		go func(m string) {
			errCh <- result{m, c.uploadOne(m, dstPath, items)}
		}(m)
	}
	var errs []string
	for range machines {
		r := <-errCh
		if r.err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", r.machine, r.err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "; "))
	}
	return nil
}

type item struct{ rel, abs string }

func (c *Config) uploadOne(machine, dstPath string, items []item) error {
	for _, it := range items {
		// 远端固定是 Linux, 用 path.Join 而非 filepath.Join
		// (Windows 上 filepath.Join 会把分隔符转成 '\')
		url := c.machineURL(machine) + "/files/file?path=" + path.Join(dstPath, filepath.ToSlash(it.rel))
		f, err := os.Open(it.abs)
		if err != nil {
			return err
		}
		st, _ := f.Stat()
		resp, err := c.do("PUT", url, f, "application/octet-stream")
		f.Close()
		if err != nil {
			return err
		}
		resp.Body.Close()
		fmt.Printf("[%s] %s (%d bytes) ✓\n", machine, filepath.ToSlash(it.rel), st.Size())
	}
	return nil
}

// parseDst: "nas1,nas2:/opt/app" → (["nas1","nas2"], "/opt/app")
func parseDst(dst string) ([]string, string, error) {
	idx := strings.Index(dst, ":")
	if idx < 0 {
		return nil, "", fmt.Errorf("目标格式应为 机器:/路径, 得到: %s", dst)
	}
	machines := strings.Split(dst[:idx], ",")
	path := dst[idx+1:]
	if path == "" {
		path = "/"
	}
	return machines, path, nil
}
