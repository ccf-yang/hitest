package server

// 端到端测试: 真实 agent 子进程 + 隧道 + 终端流 + 本机内置 agent。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"hitest/internal/agent"
	"hitest/internal/store"
	"hitest/internal/transport"
)

const e2eToken = "e2e-token"

func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	ctx := context.Background()
	st, err := store.Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	sock, err := agent.LocalStart(ctx, agent.Options{
		ID:       "local",
		SockPath: filepath.Join(t.TempDir(), "agent.sock"),
	})
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{
		cfg:       Config{Token: e2eToken},
		hub:       transport.NewHub(),
		store:     st,
		localSock: sock,
		ctx:       ctx,
	}
	ts := httptest.NewServer(s.buildGin())
	t.Cleanup(ts.Close)
	return ts
}

func startAgent(t *testing.T, httpURL, token, name string) *exec.Cmd {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "hitest")
	build := exec.Command("go", "build", "-o", bin, "./cmd/hitest")
	build.Dir = "../.."
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("构建: %v\n%s", err, out)
	}
	wsURL := "ws" + strings.TrimPrefix(httpURL, "http")
	cmd := exec.Command(bin, "agent", "-s", wsURL, "-t", token, "-n", name)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cmd.Process.Kill()
		cmd.Wait()
	})
	return cmd
}

func pollAgents(t *testing.T, httpURL string, want func(map[string]bool) bool, d time.Duration) {
	t.Helper()
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		resp, err := http.Get(httpURL + "/api/v1/agents")
		if err == nil {
			var res struct {
				Agents []struct {
					ID     string `json:"id"`
					Online bool   `json:"online"`
				} `json:"agents"`
			}
			json.NewDecoder(resp.Body).Decode(&res)
			resp.Body.Close()
			online := map[string]bool{}
			for _, a := range res.Agents {
				if a.Online {
					online[a.ID] = true
				}
			}
			if want(online) {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("pollAgents 超时")
}

func apiRequest(t *testing.T, method, url string, body io.Reader) (*http.Response, string) {
	t.Helper()
	resp, err := http.DefaultClient.Do(mustReq(method, url, body))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		t.Fatalf("%s %s → HTTP %d: %s", method, url, resp.StatusCode, data)
	}
	return resp, string(data)
}

func mustReq(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		panic(err)
	}
	return req
}

// ---------- 1. 隧道 API: 文件三连 ----------

func TestE2E_TunnelFiles(t *testing.T) {
	ts := newTestServer(t)
	startAgent(t, ts.URL, e2eToken, "e2e")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return on["e2e"] }, 15*time.Second)

	base := ts.URL + "/api/v1/m/e2e"
	shared := t.TempDir() // 与 agent 同机

	// 上传（流式 PUT）
	content := bytes.Repeat([]byte("hitest-e2e-"), 10000)
	_, respBody := apiRequest(t, "PUT", base+"/files/file?path="+shared+"/a.bin", bytes.NewReader(content))
	if !strings.Contains(respBody, `"ok"`) {
		t.Fatalf("上传应答异常: %s", respBody)
	}
	got, _ := os.ReadFile(shared + "/a.bin")
	if !bytes.Equal(got, content) {
		t.Fatalf("上传内容不一致: %d vs %d bytes", len(got), len(content))
	}

	// 列目录
	_, listBody := apiRequest(t, "GET", base+"/files/list?path="+shared, nil)
	if !strings.Contains(listBody, `"a.bin"`) {
		t.Fatalf("列目录缺 a.bin: %s", listBody)
	}

	// 下载（Range 兼容的 ServeContent）
	resp, err := http.Get(base + "/files/file?path=" + shared + "/a.bin")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	downloaded, _ := io.ReadAll(resp.Body)
	if !bytes.Equal(downloaded, content) {
		t.Error("下载内容不一致")
	}

	// 删除
	_, delBody := apiRequest(t, "POST", base+"/files/delete",
		strings.NewReader(`{"path":"`+shared+`/a.bin"}`))
	if !strings.Contains(delBody, `"ok"`) {
		t.Fatalf("删除失败: %s", delBody)
	}
	if _, err := os.Stat(shared + "/a.bin"); !os.IsNotExist(err) {
		t.Error("a.bin 应已删除")
	}
}

// ---------- 2. 隧道 API: exec（AI agent 工具形态） ----------

func TestE2E_TunnelExec(t *testing.T) {
	ts := newTestServer(t)
	startAgent(t, ts.URL, e2eToken, "e2e")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return on["e2e"] }, 15*time.Second)

	body := fmt.Sprintf(`{"cmd":"echo e2e-exec-$((40+2))","timeout_sec":10}`)
	_, respBody := apiRequest(t, "POST", ts.URL+"/api/v1/m/e2e/exec/run", strings.NewReader(body))
	var res struct {
		Code   int    `json:"code"`
		Stdout string `json:"stdout"`
	}
	if err := json.Unmarshal([]byte(respBody), &res); err != nil {
		t.Fatalf("应答解析: %v (%s)", err, respBody)
	}
	if res.Code != 0 || !strings.Contains(res.Stdout, "e2e-exec-42") {
		t.Errorf("exec 异常: %+v", res)
	}

	// 失败命令返回非零 exit code
	_, respBody = apiRequest(t, "POST", ts.URL+"/api/v1/m/e2e/exec/run",
		strings.NewReader(`{"cmd":"exit 7"}`))
	var res2 struct {
		Code int `json:"code"`
	}
	json.Unmarshal([]byte(respBody), &res2)
	if res2.Code != 7 {
		t.Errorf("exit code = %d, want 7", res2.Code)
	}
}

// ---------- 3. 终端流 ----------

func TestE2E_Terminal(t *testing.T) {
	ts := newTestServer(t)
	startAgent(t, ts.URL, e2eToken, "e2e")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return on["e2e"] }, 15*time.Second)

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/api/v1/m/e2e/terminal/ws"
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()

	open, _ := json.Marshal(map[string]any{"cols": 100, "rows": 30, "cmd": "/bin/bash -l"})
	ws.WriteMessage(websocket.TextMessage, open)

	// 键入 echo（帧: [0][payload]）
	time.Sleep(500 * time.Millisecond) // 等 pty 就绪
	input := append([]byte{0}, []byte("echo hitest-term-99\n")...)
	ws.WriteMessage(websocket.BinaryMessage, input)

	// 收输出直到出现目标
	deadline := time.Now().Add(15 * time.Second)
	var buf bytes.Buffer
	for time.Now().Before(deadline) {
		ws.SetReadDeadline(time.Now().Add(2 * time.Second))
		mt, data, err := ws.ReadMessage()
		if err != nil {
			break
		}
		if mt == websocket.BinaryMessage && len(data) > 0 && data[0] == 0 {
			buf.Write(data[1:])
			if bytes.Contains(buf.Bytes(), []byte("hitest-term-99")) {
				// 4. resize 生效检查
				ws.WriteMessage(websocket.BinaryMessage, append([]byte{1}, []byte{0, 80, 0, 24}...))
				time.Sleep(300 * time.Millisecond)
				ws.WriteMessage(websocket.BinaryMessage, append([]byte{0}, []byte("stty size\n")...))
				if waitSubstring(&buf, "24 80", 10*time.Second, ws) {
					return // PASS
				}
				t.Fatalf("stty size 未生效, 已收: %q", buf.String())
			}
		}
	}
	t.Fatalf("终端输出超时, 已收: %q", buf.String())
}

func waitSubstring(buf *bytes.Buffer, needle string, d time.Duration, ws *websocket.Conn) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if bytes.Contains(buf.Bytes(), []byte(needle)) {
			return true
		}
		ws.SetReadDeadline(time.Now().Add(1 * time.Second))
		mt, data, err := ws.ReadMessage()
		if err != nil {
			continue
		}
		if mt == websocket.BinaryMessage && len(data) > 0 && data[0] == 0 {
			buf.Write(data[1:])
		}
	}
	return bytes.Contains(buf.Bytes(), []byte(needle))
}

// ---------- 4. 本机内置 agent ----------

func TestE2E_LocalBuiltin(t *testing.T) {
	ts := newTestServer(t)
	_, body := apiRequest(t, "GET", ts.URL+"/api/v1/local/monitor/base/os", nil)
	if !strings.Contains(body, `"platform"`) {
		t.Errorf("本机 monitor/base/os 异常: %s", body)
	}
	_, body = apiRequest(t, "GET", ts.URL+"/api/v1/local/files/list?path=/", nil)
	if !strings.Contains(body, `"entries"`) {
		t.Errorf("本机 files/list 异常: %s", body)
	}
	// 移植模块冒烟: toolbox / disks / monitor 采样
	_, body = apiRequest(t, "POST", ts.URL+"/api/v1/local/toolbox/device/base", strings.NewReader(`{}`))
	if !strings.Contains(body, `"hostname"`) || !strings.Contains(body, "swapMemoryTotal") {
		t.Errorf("toolbox/device/base 异常: %s", body)
	}
	_, body = apiRequest(t, "GET", ts.URL+"/api/v1/local/disks", nil)
	if !strings.Contains(body, "totalDisks") {
		t.Errorf("disks 异常: %s", body)
	}
}

// ---------- 5. 异常路径 ----------

func TestE2E_OfflineAgent(t *testing.T) {
	ts := newTestServer(t)
	resp, err := http.Get(ts.URL + "/api/v1/m/ghost/files/list")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 503 {
		t.Errorf("离线 agent 应返回 503, 得到 %d", resp.StatusCode)
	}
}

func TestE2E_WrongTokenRejected(t *testing.T) {
	ts := newTestServer(t)
	startAgent(t, ts.URL, "wrong-token", "intruder")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return len(on) == 0 }, 5*time.Second)
}

func TestE2E_DuplicateAgentID(t *testing.T) {
	ts := newTestServer(t)
	startAgent(t, ts.URL, e2eToken, "dup")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return on["dup"] }, 15*time.Second)
	startAgent(t, ts.URL, e2eToken, "dup")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return on["dup"] }, 15*time.Second)
	// server 仍健康
	apiRequest(t, "GET", ts.URL+"/api/v1/local/monitor/base/os", nil)
}

// ---------- 6. 信封中间件（Flutter 客户端兼容层） ----------

func TestE2E_Envelope(t *testing.T) {
	ts := newTestServer(t)
	startAgent(t, ts.URL, e2eToken, "e2e")
	pollAgents(t, ts.URL, func(on map[string]bool) bool { return on["e2e"] }, 15*time.Second)

	// 带信封头 → {code:200, data:{...}}（legacy 客户端解析约定）
	req, _ := http.NewRequest("GET", ts.URL+"/api/v1/m/e2e/monitor/base/os", nil)
	req.Header.Set("X-Hitest-Envelope", "1")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var env struct {
		Code    int            `json:"code"`
		Message string         `json:"message"`
		Data    map[string]any `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil {
		t.Fatalf("信封解析: %v", err)
	}
	if env.Code != 200 || env.Data["os"] == nil {
		t.Errorf("信封异常: %+v", env)
	}

	// 不带头 → 裸 JSON（web 前端约定）
	resp2, err := http.Get(ts.URL + "/api/v1/m/e2e/monitor/base/os")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	var bare map[string]any
	json.NewDecoder(resp2.Body).Decode(&bare)
	if _, ok := bare["code"]; ok {
		t.Errorf("裸响应不应有 code 字段: %v", bare)
	}
	if bare["os"] == nil {
		t.Errorf("裸响应缺字段: %v", bare)
	}
}
