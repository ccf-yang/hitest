package ssh

import (
	"errors"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

type ConnInfo struct {
	User       string `json:"user"`
	Addr       string `json:"addr"`
	Port       int    `json:"port"`
	AuthMode   string `json:"authMode"`
	Password   string `json:"password"`
	PrivateKey []byte `json:"privateKey"`
	PassPhrase []byte `json:"passPhrase"`

	UseProxy    bool          `json:"useProxy"`
	DialTimeOut time.Duration `json:"dialTimeOut"`
}

type SSHClient struct {
	Client   *gossh.Client `json:"client"`
	SudoItem string        `json:"sudoItem"`
}

var ErrCommandTerminationUnconfirmed = errors.New("remote command termination could not be confirmed; check the source node before retrying")

type streamCallbackWriter struct {
	mu       sync.Mutex
	callback func(string)
}
