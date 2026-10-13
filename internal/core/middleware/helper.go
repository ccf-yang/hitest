package middleware

import (
	"strings"

	"hitest/internal/core/utils/publicshare"
)

func ShouldProxyToAgent(reqPath string) bool {
	if strings.HasPrefix(reqPath, "/hitest/swagger") || !strings.HasPrefix(reqPath, "/api/v2") {
		return false
	}
	if strings.HasPrefix(reqPath, "/api/v2/core") && !strings.HasPrefix(reqPath, "/api/v2/core/xpack") {
		return false
	}
	return true
}

func IsPublicFileShareAPI(reqPath string) bool {
	return publicshare.IsAPI(reqPath)
}
