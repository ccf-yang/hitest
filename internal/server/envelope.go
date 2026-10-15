package server

// envelope 中间件: 为 Flutter 客户端(带 X-Hitest-Envelope: 1 头)提供
// legacy 的 {code, message, data} 响应信封, 使客户端既有解析代码零改动。
// Web 前端不带此头, 继续拿裸 JSON。
//
// 规则:
//   - 仅作用于 /api/v1 下的 JSON API; 文件流(/files/)与 WS(/ws) 不包裹
//   - HTTP 200        → {code:200, message:"success", data:<原响应>}
//   - HTTP 4xx/5xx    → HTTP 200 + {code:<原状态码>, message:<error>, data:<原响应>}
//     (与 1Panel 客户端的错误处理约定一致)

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const envelopeHeader = "X-Hitest-Envelope"

type bufferedWriter struct {
	gin.ResponseWriter
	buf bytes.Buffer
}

func (b *bufferedWriter) Write(data []byte) (int, error)       { return b.buf.Write(data) }
func (b *bufferedWriter) WriteString(s string) (int, error)    { return b.buf.WriteString(s) }
func (b *bufferedWriter) WriteHeaderNow()                      { b.ResponseWriter.WriteHeaderNow() }

func (s *Server) envelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader(envelopeHeader) != "1" {
			c.Next()
			return
		}
		p := c.Request.URL.Path
		if !strings.HasPrefix(p, "/api/v1/") ||
			strings.Contains(p, "/files/") ||
			strings.HasSuffix(p, "/ws") {
			c.Next()
			return
		}

		bw := &bufferedWriter{ResponseWriter: c.Writer}
		c.Writer = bw
		c.Next()

		// 还原真实 writer, 重写响应
		c.Writer = bw.ResponseWriter
		status := bw.Status()
		c.Writer.Header().Del("Content-Length")

		var payload any
		if bw.buf.Len() > 0 && strings.Contains(bw.Header().Get("Content-Type"), "json") {
			_ = json.Unmarshal(bw.buf.Bytes(), &payload)
		}
		if payload == nil {
			payload = gin.H{}
		}

		if status >= 200 && status < 300 {
			c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": payload})
			return
		}
		// 错误: 保持 200 外壳 + 业务 code, 兼容 legacy 客户端约定
		msg := "error"
		if m, ok := payload.(map[string]any); ok {
			if e, ok := m["error"].(string); ok && e != "" {
				msg = e
			}
		}
		c.JSON(http.StatusOK, gin.H{"code": status, "message": msg, "data": payload})
	}
}
