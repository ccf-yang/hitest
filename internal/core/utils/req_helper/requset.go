package req_helper

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"hitest/internal/core/buserr"
)

func HandleGet(url string) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: (&net.Dialer{
			Timeout:   60 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       15 * time.Second,
	}
	return handleGetWithTransport(url, transport)
}

func handleGetWithTransport(url string, transport *http.Transport) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 300,
	}
	client.Transport = transport
	defer client.CloseIdleConnections()

	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, buserr.WithMap("ErrCreateHttpClient", map[string]interface{}{"err": err.Error()}, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, buserr.WithMap("ErrHttpReqTimeOut", map[string]interface{}{"err": err.Error()}, err)
		} else {
			if strings.Contains(err.Error(), "no such host") {
				return nil, buserr.WithErr("ErrNoSuchHost", errors.New("no such host"))
			}
			return nil, buserr.WithMap("ErrHttpReqFailed", map[string]interface{}{"err": err.Error()}, err)
		}
	}
	if resp.StatusCode == 404 {
		_ = resp.Body.Close()
		return nil, buserr.WithErr("ErrHttpReqNotFound", errors.New("no such resource"))
	}

	return resp, nil
}
