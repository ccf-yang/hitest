package providers

import (
	"net/http"

	"hitest/internal/shared/model"

	"github.com/gin-gonic/gin"
)

type MultiNodeProvider interface {
	IsXpack() bool
	IsUseCustomApp() bool
	GetImagePrefix() string
	LoadNodeInfo(isBase bool) (model.NodeInfo, error)

	LoadRequestTransport() *http.Transport
	ValidateCertificate(c *gin.Context) bool
}
