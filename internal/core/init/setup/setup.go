package setup

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"hitest/internal/shared/ctl_conf"

	"github.com/gin-gonic/gin"
)

//go:embed setup.html
var setupFS embed.FS

// NeedSetup 判断是否需要初始化：配置文件不存在即为未初始化。
func NeedSetup() bool {
	_, err := os.Stat(configFile())
	return os.IsNotExist(err)
}

func configFile() string {
	baseDir := ctl_conf.Load("BASE_DIR")
	if baseDir == "" {
		baseDir = "/opt"
	}
	return path.Join(baseDir, "hitest", "conf", "app.yaml")
}

type initRequest struct {
	BaseDir  string `json:"baseDir"`
	Port     string `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Run 启动一个极简的 Web 初始化服务，用户提交后写入配置文件并返回，
// 随后调用方继续正常的初始化流程（不重启进程）。
func Run() {
	port := os.Getenv("HITEST_PORT")
	if port == "" {
		port = ctl_conf.Load("ORIGINAL_PORT")
	}
	if port == "" {
		port = "9999"
	}

	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) {
		html, _ := setupFS.ReadFile("setup.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", html)
	})
	engine.GET("/api/v2/core/setup/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"initialized": !NeedSetup()})
	})

	done := make(chan struct{})
	engine.POST("/api/v2/core/setup/init", func(c *gin.Context) {
		var req initRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求格式错误"})
			return
		}
		if err := apply(req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
		go func() {
			time.Sleep(300 * time.Millisecond)
			close(done)
		}()
	})

	server := &http.Server{Addr: ":" + port, Handler: engine}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[setup] listen on :%s failed: %v\n", port, err)
		}
	}()

	fmt.Printf("\n========================================\n")
	fmt.Printf(" hitest 未初始化，请在浏览器打开：\n")
	fmt.Printf("   http://<本机IP>:%s\n", port)
	fmt.Printf(" 完成初始化后将自动继续启动面板。\n")
	fmt.Printf("========================================\n\n")

	<-done
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	fmt.Println("[setup] 初始化完成，继续启动面板...")
}

func apply(req initRequest) error {
	baseDir := strings.TrimSpace(req.BaseDir)
	if baseDir == "" {
		baseDir = "/opt"
	}
	if !path.IsAbs(baseDir) {
		return fmt.Errorf("安装目录必须是绝对路径")
	}
	port := strings.TrimSpace(req.Port)
	if port == "" {
		port = "9999"
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		username = "admin"
	}
	if strings.TrimSpace(req.Password) == "" {
		return fmt.Errorf("管理员密码不能为空")
	}

	for _, dir := range []string{"conf", "db", "log"} {
		if err := os.MkdirAll(path.Join(baseDir, "hitest", dir), 0o755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	content := fmt.Sprintf(`base:
  install_dir: %s
  mode: dev
  username: %s
  password: %s
  language: zh
  version: v2.0.0
conn:
  port: "%s"
  bindAddress: ""
  ssl: ""
  ipv6: ""
  entrance: ""
`, baseDir, username, req.Password, port)

	confFile := path.Join(baseDir, "hitest", "conf", "app.yaml")
	if err := os.WriteFile(confFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入配置失败: %w", err)
	}

	// 写入安装目录指针，供后续启动在没有 hitestctl / 环境变量时定位 BASE_DIR。
	if err := os.MkdirAll("/etc/hitest", 0o755); err == nil {
		_ = os.WriteFile("/etc/hitest/base_dir", []byte(baseDir+"\n"), 0o644)
	}

	// 让同一进程内后续的配置读取（ctl_conf / viper）立刻拿到新的 BASE_DIR。
	_ = os.Setenv("HITEST_BASE_DIR", baseDir)
	return nil
}
