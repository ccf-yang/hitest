package common

import (
	"crypto/rand"
	"math/big"
	"net"
	"os"
	"path"

	"github.com/gin-gonic/gin"

	"hitest/internal/core/global"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func RandStrAndNum(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	max := big.NewInt(int64(len(charset)))
	for i := range b {
		num, err := rand.Int(rand.Reader, max)
		if err != nil {
			return ""
		}
		b[i] = charset[num.Int64()]
	}
	return string(b)
}

func CreateDirWhenNotExist(isDir bool, pathItem string) (string, error) {
	checkPath := pathItem
	if !isDir {
		checkPath = path.Dir(pathItem)
	}
	if _, err := os.Stat(checkPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(checkPath, os.ModePerm); err != nil {
			global.LOG.Errorf("mkdir %s failed, err: %v", checkPath, err)
			return pathItem, err
		}
	}
	return pathItem, nil
}

func CheckIpInCidr(cidr, checkIP string) bool {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		global.LOG.Errorf("parse CIDR %s failed, err: %v", cidr, err)
		return false
	}
	return ipNet.Contains(net.ParseIP(checkIP))
}

func GetRealClientIP(c *gin.Context) string {
	addr := c.Request.RemoteAddr
	if ip, _, err := net.SplitHostPort(addr); err == nil {
		return ip
	}
	return addr
}
