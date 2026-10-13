package security

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"hitest/internal/core/app/repo"
	"hitest/internal/core/cmd/server/res"
	"hitest/internal/core/utils/common"
	"hitest/internal/shared/constant"
	"hitest/web"

	"github.com/gin-gonic/gin"
)

func HandleNotRoute(c *gin.Context) bool {
	if !checkBindDomain(c) {
		HandleNotSecurity(c, "err_domain")
		return false
	}
	if !checkIPLimit(c) {
		HandleNotSecurity(c, "err_ip_limit")
		return false
	}
	if CanServeFrontendPath(c) {
		ToIndexHtml(c)
		return false
	}
	return true
}

// CheckSecurity 免登录面板：仅做绑定域名与 IP 白名单校验。
func CheckSecurity(c *gin.Context) bool {
	if !checkBindDomain(c) {
		HandleNotSecurity(c, "err_domain")
		return false
	}
	if !checkIPLimit(c) {
		HandleNotSecurity(c, "err_ip_limit")
		return false
	}
	return true
}

func ToIndexHtml(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	data, err := web.IndexHtml.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html not found")
		return
	}
	_, _ = c.Writer.Write(data)
	c.Writer.Flush()
}

func HandleNotSecurity(c *gin.Context, resType string) {
	code := LoadErrCode()
	if code == 444 {
		CloseDirectly(c)
		return
	}

	file := fmt.Sprintf("html/%d.html", code)
	if code == http.StatusOK && resType != "" {
		file = fmt.Sprintf("html/200_%s.html", resType)
	}
	data, err := res.ErrorMsg.ReadFile(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	c.Data(code, "text/html; charset=utf-8", data)
}

func LoadErrCode() int {
	settingRepo := repo.NewISettingRepo()
	codeVal, err := settingRepo.GetValueByKey("NoAuthSetting")
	if err != nil {
		return http.StatusInternalServerError
	}

	switch codeVal {
	case "400":
		return http.StatusBadRequest
	case "401":
		return http.StatusUnauthorized
	case "403":
		return http.StatusForbidden
	case "404":
		return http.StatusNotFound
	case "408":
		return http.StatusRequestTimeout
	case "416":
		return http.StatusRequestedRangeNotSatisfiable
	case "500":
		return http.StatusInternalServerError
	case "444":
		return 444
	default:
		return http.StatusOK
	}
}

func IsFrontendPath(path string) bool {
	reqUri := strings.TrimSuffix(path, "/")
	if _, ok := constant.WebUrlMap[reqUri]; ok {
		return true
	}
	for _, route := range constant.DynamicRoutes {
		if match, _ := regexp.MatchString(route, reqUri); match {
			return true
		}
	}
	return false
}

func CanServeFrontendPath(c *gin.Context) bool {
	return IsFrontendPath(c.Request.URL.Path)
}

func checkBindDomain(c *gin.Context) bool {
	settingRepo := repo.NewISettingRepo()
	status, _ := settingRepo.Get(repo.WithByKey("BindDomain"))
	if len(status.Value) == 0 {
		return true
	}
	domains := c.Request.Host
	parts := strings.Split(c.Request.Host, ":")
	if len(parts) > 0 {
		domains = parts[0]
	}
	return domains == status.Value
}

func checkIPLimit(c *gin.Context) bool {
	settingRepo := repo.NewISettingRepo()
	status, _ := settingRepo.Get(repo.WithByKey("AllowIPs"))
	if len(status.Value) == 0 {
		return true
	}
	trustedProxies, _ := settingRepo.Get(repo.WithByKey("AllowIPTrustedProxies"))
	clientIP := common.ResolveClientIP(c, trustedProxies.Value)

	for _, ip := range strings.Split(status.Value, ",") {
		if len(ip) == 0 {
			continue
		}
		if ip == clientIP || (strings.Contains(ip, "/") && common.CheckIpInCidr(ip, clientIP)) {
			return true
		}
	}
	return false
}

func CloseDirectly(c *gin.Context) {
	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	conn.Close()
	c.Abort()
}
