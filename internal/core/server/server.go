package server

import (
	"bufio"
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"hitest/internal/core/init/db"
	"hitest/internal/core/init/geo"
	"hitest/internal/core/init/log"
	"hitest/internal/core/init/migration"
	"hitest/internal/core/init/proxy"

	"github.com/gin-gonic/gin"
	"github.com/soheilhy/cmux"

	agentServer "hitest/internal/agent/server"

	"hitest/internal/core/global"
	"hitest/internal/core/i18n"
	"hitest/internal/core/init/cron"
	"hitest/internal/core/init/hook"
	"hitest/internal/core/init/router"
	"hitest/internal/core/init/session"
	"hitest/internal/core/init/session/psession"
	"hitest/internal/core/init/setup"
	"hitest/internal/core/init/validator"
	"hitest/internal/core/init/viper"
	"hitest/internal/shared/constant"
	"hitest/internal/shared/re"
)

func Start() {
	re.Init()
	if setup.NeedSetup() {
		setup.Run()
	}
	viper.Init()
	log.Init()
	db.Init()
	migration.Init()
	i18n.Init()
	validator.Init()
	geo.Init()
	gob.Register(psession.SessionUser{})
	cron.Init()
	session.Init()
	InitOthersBeforeHook()
	hook.Init()
	InitOthersAfterHook()

	proxy.Init()

	go agentServer.Start()
	waitForAgentSocket()

	rootRouter := router.Routers()

	if global.CONF.Base.Mode != "stable" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	tcpItem := "tcp4"
	if global.CONF.Conn.Ipv6 == constant.StatusEnable {
		tcpItem = "tcp"
		global.CONF.Conn.BindAddress = fmt.Sprintf("[%s]", global.CONF.Conn.BindAddress)
	}
	server := &http.Server{
		Addr:              global.CONF.Conn.BindAddress + ":" + global.CONF.Conn.Port,
		Handler:           rootRouter,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       600 * time.Second,
		WriteTimeout:      600 * time.Second,
		IdleTimeout:       240 * time.Second,
	}
	ln, err := net.Listen(tcpItem, server.Addr)
	if err != nil {
		panic(err)
	}
	type tcpKeepAliveListener struct {
		*net.TCPListener
	}
	switch global.CONF.Conn.SSL {
	case constant.StatusEnable:
		constant.CertStore.Store(loadCert())

		server.TLSConfig = &tls.Config{
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return constant.CertStore.Load().(*tls.Certificate), nil
			},
		}
		global.LOG.Infof("listen at https://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)

		if err := server.ServeTLS(tcpKeepAliveListener{ln.(*net.TCPListener)}, "", ""); err != nil {
			panic(err)
		}
	case constant.StatusMux:
		constant.CertStore.Store(loadCert())

		server.TLSConfig = &tls.Config{
			NextProtos: []string{"h2", "http/1.1"},
			GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
				return constant.CertStore.Load().(*tls.Certificate), nil
			},
		}

		m := cmux.New(ln)

		httpsL := m.Match(cmux.TLS())
		httpL := m.Match(cmux.HTTP1Fast())
		anyL := m.Match(cmux.Any())

		go func() {
			if err := server.Serve(tls.NewListener(httpsL, server.TLSConfig)); err != nil {
				global.LOG.Errorf("HTTPS Serve Error: %v", err)
			}
		}()

		go func() {
			for {
				conn, err := httpL.Accept()
				if err != nil {
					return
				}
				go handleMuxHttpConn(conn)
			}
		}()

		go func() {
			for {
				conn, err := anyL.Accept()
				if err != nil {
					return
				}
				conn.Close()
			}
		}()

		global.LOG.Infof("listen at mux (http/https)://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)

		if err := m.Serve(); err != nil {
			panic(err)
		}
	default:
		global.LOG.Infof("listen at http://%s:%s [%s]", global.CONF.Conn.BindAddress, global.CONF.Conn.Port, tcpItem)
		if err := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)}); err != nil {
			panic(err)
		}
	}
}

func loadCert() *tls.Certificate {
	certPath := path.Join(global.CONF.Base.InstallDir, "hitest/secret/server.crt")
	keyPath := path.Join(global.CONF.Base.InstallDir, "hitest/secret/server.key")
	certificate, err := os.ReadFile(certPath)
	if err != nil {
		panic(err)
	}
	key, err := os.ReadFile(keyPath)
	if err != nil {
		panic(err)
	}
	cert, err := tls.X509KeyPair(certificate, key)
	if err != nil {
		panic(err)
	}
	return &cert
}

func handleMuxHttpConn(conn net.Conn) {
	defer conn.Close()

	req, err := http.ReadRequest(bufio.NewReader(conn))
	if err != nil {
		return
	}

	if req.Host == "" {
		return
	}

	ua := req.Header.Get("User-Agent")
	if ua == "" {
		return
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost:
	default:
		return
	}

	if len(req.RequestURI) > 4096 {
		return
	}

	if !strings.HasPrefix(req.URL.Path, "/") {
		return
	}

	target := "https://" + req.Host + req.URL.RequestURI()

	resp := &http.Response{
		StatusCode: http.StatusTemporaryRedirect,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}
	resp.Header.Set("Location", target)
	resp.Header.Set("Connection", "close")

	_ = resp.Write(conn)
}

// waitForAgentSocket blocks until the embedded agent has created its unix socket,
// so that the first proxied request does not race with agent startup.
func waitForAgentSocket() {
	deadline := time.Now().Add(60 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(proxy.SockPath); err == nil {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	global.LOG.Warnf("embedded agent socket %s not ready after timeout", proxy.SockPath)
}
