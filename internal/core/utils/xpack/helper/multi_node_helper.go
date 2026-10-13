package helper

import (
	"net/http"

	"hitest/internal/core/global"
	"hitest/internal/core/init/proxy"

	"github.com/gin-gonic/gin"
)

type multiNodeHelper struct{}

func NewIMultiNodeProvider() *multiNodeHelper {
	return &multiNodeHelper{}
}

func (m *multiNodeHelper) Proxy(c *gin.Context, currentNode string) {
	if currentNode != "local" && currentNode != "" {
		c.Next()
		return
	}
	defer func() {
		if err := recover(); err != nil && err != http.ErrAbortHandler {
			global.LOG.Debug(err)
		}
	}()
	proxy.LocalAgentProxy.ServeHTTP(c.Writer, c.Request)
	c.Abort()
}
