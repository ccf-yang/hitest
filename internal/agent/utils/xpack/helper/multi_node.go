package helper

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"hitest/internal/agent/global"
	"hitest/internal/agent/utils/common"
	"hitest/internal/agent/utils/xpack/providers"
	"hitest/internal/shared/model"

	"github.com/gin-gonic/gin"
)

type multiNodeHelper struct{}

func NewIMultiNodeProvider() providers.MultiNodeProvider {
	return &multiNodeHelper{}
}

func (m *multiNodeHelper) LoadNodeInfo(isBase bool) (model.NodeInfo, error) {
	var info model.NodeInfo
	info.BaseDir = common.LoadParams("BASE_DIR")
	info.Version = common.LoadParams("ORIGINAL_VERSION")
	info.Scope = "master"
	global.IsMaster = true
	return info, nil
}

func (m *multiNodeHelper) GetImagePrefix() string {
	return ""
}

func (m *multiNodeHelper) IsUseCustomApp() bool {
	return false
}

func (m *multiNodeHelper) IsXpack() bool {
	return false
}

func (m *multiNodeHelper) LoadRequestTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
}

func (m *multiNodeHelper) ValidateCertificate(c *gin.Context) bool {
	return true
}
