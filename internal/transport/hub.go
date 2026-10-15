package transport

import (
	"sync"
)

// Register 是 agent 上行注册消息（WS 连接建立后的第一条文本消息）
type Register struct {
	ID       string `json:"id"`       // 机器 ID（如 nas1, armbian-1）
	Token    string `json:"token"`
	Hostname string `json:"hostname"`
	OS       string `json:"os"` // linux/arm64
}

type RegisterAck struct {
	OK      bool   `json:"ok"`
	Error   string `json:"error,omitempty"`
}

// Hub 是 server 侧的在线 agent 注册表。
// 只存"谁在线 + 它的 yamux 会话"，不存历史（历史在 store）。
type Hub struct {
	mu     sync.Mutex
	agents map[string]*Session
}

func NewHub() *Hub {
	return &Hub{agents: make(map[string]*Session)}
}

// Register 登记一个 agent；同 ID 旧会话被顶掉（返回旧会话供调用方关闭）
func (h *Hub) Register(id string, s *Session) (old *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	old = h.agents[id]
	h.agents[id] = s
	return old
}

func (h *Hub) Unregister(id string, s *Session) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.agents[id] == s {
		delete(h.agents, id)
	}
}

func (h *Hub) Get(id string) *Session {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.agents[id]
}

func (h *Hub) IDs() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	ids := make([]string, 0, len(h.agents))
	for id := range h.agents {
		ids = append(ids, id)
	}
	return ids
}

func (h *Hub) Online(id string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.agents[id]
	return s != nil && !s.IsClosed()
}
