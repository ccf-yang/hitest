// Package server: serve 角色。
//
//   /agent/ws                  agent 接入点（注册 → yamux）
//   /api/v1/agents             机器列表（在线 + 历史注册）
//   /api/v1/local/*            反代到内置本机 agent（unix socket）
//   /api/v1/m/{id}/*           隧道反代到远程 agent（stream[http]）
//     └ 其中 */terminal/ws     终端（本机直连 pty / 远程走 stream[term]）
//   /                          内嵌前端
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"strings"
	"net/url"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"hitest/internal/agent"
	"hitest/internal/modules/terminal"
	"hitest/internal/store"
	"hitest/internal/transport"
)

type Config struct {
	Port        int    `yaml:"port"`
	Token       string `yaml:"token"`         // agent 注册用
	AccessToken string `yaml:"access_token"`  // Web/API 访问用（空 = 局域网无鉴权）
	CertFile    string `yaml:"cert_file"`     // TLS（简单模式）
	KeyFile     string `yaml:"key_file"`
	DataDir     string `yaml:"data_dir"`
}

type Server struct {
	cfg   Config
	hub   *transport.Hub
	store *store.Store
	localSock string
	ctx   context.Context
}

func Run(ctx context.Context, cfg Config) error {
	if cfg.Port == 0 {
		cfg.Port = 8080
	}
	if cfg.DataDir == "" {
		cfg.DataDir = "data"
	}
	if cfg.Token == "" {
		return fmt.Errorf("config.token 不能为空（首次部署请手动生成一个随机串）")
	}
	st, err := store.Open(filepath.Join(cfg.DataDir, "hitest.db"))
	if err != nil {
		return fmt.Errorf("打开数据库: %w", err)
	}

	s := &Server{cfg: cfg, hub: transport.NewHub(), store: st, ctx: ctx}

	// 内置本机 agent（同一进程, 经 unix socket, 与远程 agent 完全同构）
	s.localSock, err = agent.LocalStart(ctx, agent.Options{
		ID:       "local",
		SockPath: filepath.Join(cfg.DataDir, "agent-local.sock"),
		DataDir:  cfg.DataDir,
	})
	if err != nil {
		return err
	}
	return s.listen()
}

// buildGin 组装路由
func (s *Server) buildGin() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	if s.cfg.AccessToken != "" {
		r.Use(s.requireAccessToken)
	}
	r.Use(s.envelope())

	r.GET("/agent/ws", s.handleAgentWS)
	r.GET("/api/v1/agents", s.handleAgents)

	// 本机 + 远程统一为 machine 维度; "local" = 内置 agent
	r.Any("/api/v1/local/*rest", s.routeLocal)
	r.Any("/api/v1/m/:id/*rest", s.routeRemote)

	// 内嵌 Vue SPA
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(404, gin.H{"error": "no such api"})
			return
		}
		StaticHandler().ServeHTTP(c.Writer, c.Request)
	})
	return r
}

func (s *Server) requireAccessToken(c *gin.Context) {
	if c.Request.URL.Path == "/agent/ws" {
		return
	}
	if c.GetHeader("X-Token") != s.cfg.AccessToken {
		c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
		return
	}
}

// ---------- agent 接入 ----------

func (s *Server) handleAgentWS(c *gin.Context) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, data, err := ws.ReadMessage()
	if err != nil {
		return
	}
	var reg transport.Register
	if json.Unmarshal(data, &reg) != nil || reg.ID == "" {
		return
	}
	if reg.Token != s.cfg.Token {
		ws.WriteJSON(transport.RegisterAck{OK: false, Error: "token 不正确"})
		log.Printf("[server] agent 注册被拒绝: %s", reg.ID)
		return
	}
	ws.WriteJSON(transport.RegisterAck{OK: true})
	ws.SetReadDeadline(time.Time{})
	log.Printf("[server] agent 上线: %s (%s, %s)", reg.ID, reg.Hostname, reg.OS)
	s.store.TouchMachine(reg.ID, reg.Hostname, reg.OS)

	sess, err := transport.NewServerSession(transport.NewWSNetConn(ws))
	if err != nil {
		return
	}
	if old := s.hub.Register(reg.ID, sess); old != nil {
		old.Close() // 顶掉同 ID 旧连接
	}
	defer func() {
		s.hub.Unregister(reg.ID, sess)
		log.Printf("[server] agent 离线: %s", reg.ID)
	}()

	// server 侧 Accept 循环: http 流由反代按需主动打开,
	// 这里只兜住 agent 反向开来的流（当前无此场景, 关闭即可）
	for {
		typ, conn, err := sess.Accept()
		if err != nil {
			return
		}
		conn.Close()
		_ = typ
	}
}

// ---------- 机器列表 ----------

func (s *Server) handleAgents(c *gin.Context) {
	type AgentView struct {
		ID       string `json:"id"`
		Hostname string `json:"hostname"`
		OS       string `json:"os"`
		Online   bool   `json:"online"`
	}
	online := map[string]bool{}
	for _, id := range s.hub.IDs() {
		online[id] = true
	}
	views := []AgentView{}
	for _, m := range s.store.Machines() {
		views = append(views, AgentView{m.ID, m.Hostname, m.OS, online[m.ID]})
	}
	c.JSON(200, gin.H{"agents": views})
}

// ---------- 反代 ----------

// routeLocal: /api/v1/local/*rest
func (s *Server) routeLocal(c *gin.Context) {
	rest := c.Param("rest")
	if rest == "/terminal/ws" {
		s.terminalLocal(c)
		return
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(&url.URL{Scheme: "http", Host: "local-agent"})
			// 剥机器前缀: /api/v1/local/X → agent 的 /api/v1/X
			pr.Out.URL.Path = "/api/v1" + rest
			pr.Out.URL.RawPath = ""
			pr.SetXForwarded()
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", s.localSock)
			},
		},
		ErrorHandler: jsonErrorHandler,
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

// routeRemote: /api/v1/m/:id/*rest —— 隧道反代（设计的核心支点）
func (s *Server) routeRemote(c *gin.Context) {
	id := c.Param("id")
	rest := c.Param("rest")
	if rest == "/terminal/ws" {
		s.terminalRemote(c, id)
		return
	}
	sess := s.hub.Get(id)
	if sess == nil || sess.IsClosed() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent 不在线: " + id})
		return
	}
	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(&url.URL{Scheme: "http", Host: "agent"})
			// 剥机器前缀: /api/v1/m/{id}/X → agent 的 /api/v1/X
			pr.Out.URL.Path = "/api/v1" + rest
			pr.Out.URL.RawPath = ""
			pr.SetXForwarded()
		},
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return sess.Open(transport.StreamHTTP)
			},
		},
		ErrorHandler: jsonErrorHandler,
	}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func jsonErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	w.Write([]byte(`{"error":"tunnel: ` + err.Error() + `"}`))
}

// ---------- 终端 ----------

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // 局域网信任
}

// terminalLocal: 本机 pty 直连（与远程同一条 WS 协议）
func (s *Server) terminalLocal(c *gin.Context) {
	s.terminalBridge(c, func(cols, rows uint16, cmd string) (ptyLike, error) {
		return terminal.Start(cols, rows, cmd)
	}, nil)
}

// terminalRemote: server ↔ agent 之间开 stream[term]
func (s *Server) terminalRemote(c *gin.Context, id string) {
	sess := s.hub.Get(id)
	if sess == nil || sess.IsClosed() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "agent 不在线: " + id})
		return
	}
	s.terminalBridge(c, nil, sess)
}

// ptyLike: 本机模式用的最小接口
type ptyLike interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Resize(cols, rows uint16) error
	Close()
}

// terminalBridge: 浏览器/CLI 的 WS ↔ 终端。
// 本机模式给 openPTY, 远程模式给 sess（二选一）。
func (s *Server) terminalBridge(c *gin.Context,
	openPTY func(cols, rows uint16, cmd string) (ptyLike, error),
	sess *transport.Session) {

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	// 首条消息: open 参数
	_, data, err := ws.ReadMessage()
	if err != nil {
		return
	}
	var open struct {
		Cols uint16 `json:"cols"`
		Rows uint16 `json:"rows"`
		Cmd  string `json:"cmd"`
	}
	if json.Unmarshal(data, &open) != nil {
		return
	}
	if open.Cols == 0 {
		open.Cols = 80
	}
	if open.Rows == 0 {
		open.Rows = 24
	}

	if openPTY != nil {
		// ---- 本机模式 ----
		p, err := openPTY(open.Cols, open.Rows, open.Cmd)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
			return
		}
		defer p.Close()
		go func() {
			buf := make([]byte, 32*1024)
			for {
				n, err := p.Read(buf)
				if n > 0 {
					if werr := ws.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
						return
					}
				}
				if err != nil {
					ws.WriteMessage(websocket.BinaryMessage, []byte{terminal.FrameClose})
					return
				}
			}
		}()
		s.pumpWS(ws, func(kind byte, payload []byte) error {
			switch kind {
			case terminal.FrameData:
				_, err := p.Write(payload)
				return err
			case terminal.FrameResize:
				if len(payload) >= 4 {
					return p.Resize(u16(payload), u16(payload[2:]))
				}
			case terminal.FrameClose:
				return errClose
			}
			return nil
		})
		return
	}

	// ---- 远程模式 ----
	stream, err := terminal.OpenRemote(sess, open.Cmd, open.Cols, open.Rows)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}
	defer stream.Close()
	go func() {
		for {
			kind, payload, err := terminal.ReadFrame(stream)
			if err != nil {
				return
			}
			if kind == terminal.FrameClose {
				ws.WriteMessage(websocket.BinaryMessage, []byte{terminal.FrameClose})
				return
			}
			if kind == terminal.FrameData {
				if werr := ws.WriteMessage(websocket.BinaryMessage, append([]byte{terminal.FrameData}, payload...)); werr != nil {
					return
				}
			}
		}
	}()
	s.pumpWS(ws, func(kind byte, payload []byte) error {
		return terminal.WriteFrame(stream, kind, payload)
	})
}

var errClose = fmt.Errorf("close requested")

// pumpWS: 读 WS 消息 → [1B kind][payload] → onFrame
func (s *Server) pumpWS(ws *websocket.Conn, onFrame func(kind byte, payload []byte) error) {
	for {
		mt, data, err := ws.ReadMessage()
		if err != nil {
			return
		}
		if mt != websocket.BinaryMessage || len(data) < 1 {
			continue
		}
		if err := onFrame(data[0], data[1:]); err != nil {
			return
		}
	}
}

func u16(b []byte) uint16 {
	return uint16(b[0])<<8 | uint16(b[1])
}

// ---------- 监听 ----------

func (s *Server) listen() error {
	r := s.buildGin()
	srv := &http.Server{Addr: fmt.Sprintf(":%d", s.cfg.Port), Handler: r}

	go func() {
		<-s.ctx.Done()
		srv.Close()
	}()

	scheme := "http"
	if s.cfg.CertFile != "" {
		scheme = "https"
	}
	log.Printf("[server] hitest 启动: %s://0.0.0.0:%d  (agent 注册 token 已配置)", scheme, s.cfg.Port)
	if s.cfg.CertFile != "" {
		return srv.ListenAndServeTLS(s.cfg.CertFile, s.cfg.KeyFile)
	}
	return srv.ListenAndServe()
}
