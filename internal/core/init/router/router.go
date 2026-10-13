package router

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	baseRepo "hitest/internal/core/app/repo"
	"hitest/internal/core/global"
	"hitest/internal/core/i18n"
	psessionUtils "hitest/internal/core/init/session/psession"
	"hitest/internal/core/middleware"
	rou "hitest/internal/core/router"
	"hitest/internal/core/utils/security"
	"hitest/web"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	Router *gin.Engine
)

func setWebStatic(rootRouter *gin.RouterGroup) {
	rootRouter.StaticFS("/public", http.FS(web.Favicon))
	rootRouter.StaticFS("/favicon.ico", http.FS(web.Favicon))
	RegisterImages(rootRouter)
	rootRouter.GET("/", func(c *gin.Context) {
		if !security.CheckSecurity(c) {
			return
		}
		security.ToIndexHtml(c)
	})
}

func Routers() *gin.Engine {
	Router = gin.New()
	Router.Use(i18n.UseI18n())
	Router.Use(middleware.WhiteAllow())
	Router.Use(middleware.BindDomain())
	Router.Use(disableAuth())

	PublicGroup := Router.Group("")
	{
		PublicGroup.Use(gzip.Gzip(gzip.DefaultCompression))
		setWebStatic(PublicGroup)
	}
	if global.CONF.Base.IsDemo {
		Router.Use(middleware.DemoHandle())
	}

	Router.Use(middleware.FrontendFallback())
	Router.Use(middleware.WebSocketOriginGuard())
	Router.Use(middleware.GlobalLoading())
	Router.Use(Proxy())

	PrivateGroup := Router.Group("/api/v2/core")
	for _, router := range rou.RouterGroupApp {
		router.InitRouter(PrivateGroup)
	}

	Router.NoRoute(func(c *gin.Context) {
		if !security.HandleNotRoute(c) {
			return
		}
		security.HandleNotSecurity(c, "")
	})

	return Router
}

func RegisterImages(rootRouter *gin.RouterGroup) {
	staticDir := filepath.Join(global.CONF.Base.InstallDir, "hitest/uploads/theme")
	rootRouter.GET("/api/v2/images/*filename", func(c *gin.Context) {
		fileName := filepath.Base(c.Param("filename"))
		filePath := filepath.Join(staticDir, fileName)
		if !strings.HasPrefix(filePath, staticDir) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		f, err := os.Open(filePath)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		content := buf[:n]
		mimeType := http.DetectContentType(buf[:n])
		if strings.Contains(string(content), "<svg") {
			mimeType = "image/svg+xml"
		}
		_, _ = f.Seek(0, io.SeekStart)
		c.Header("Content-Type", mimeType)
		_, _ = io.Copy(c.Writer, f)
	})
}

// disableAuth 注入一个已认证的管理员上下文：免登录面板。
func disableAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		name, _ := baseRepo.NewISettingRepo().GetValueByKey("UserName")
		if name == "" {
			name = "hitest"
		}
		c.Set("API_AUTH", true)
		c.Set("API_AUTH_USERNAME", name)
		c.Set(psessionUtils.GinContextSessionUserKey, psessionUtils.SessionUser{
			ID:   psessionUtils.SuperAdminSessionUserID,
			Name: name,
			Role: "ADMIN",
		})
		c.Next()
	}
}
