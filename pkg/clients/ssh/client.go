/*
 *  This file is part of PETA.
 *  Copyright (C) 2025 The PETA Authors.
 *  PETA is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  PETA is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Affero General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public License
 *  along with PETA. If not, see <https://www.gnu.org/licenses/>.
 */

package ssh

import (
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"io"
	"net"
	"peta.io/peta/pkg/log"
	"peta.io/peta/pkg/utils/iputils"
	"strings"
	"sync"
	"time"
)

// DefaultTimeout is the timeout of ssh client connection.
const (
	DefaultSSHPort = 22
	DefaultTimeout = 20 * time.Second
)

// Client for ssh.
type Client struct {
	mu sync.Mutex
	*ssh.Client
	Config *Config
}

// Config for SSH Client.
type Config struct {
	User           string
	Addr           string
	Port           uint
	Auth           Auth
	Timeout        time.Duration
	Callback       ssh.HostKeyCallback
	BannerCallback ssh.BannerCallback
}

// New starts a new ssh connection, the host public key must be in known hosts.
func New(
	user, addr string,
	port uint,
	passwd, privateKey, privateKeyRaw, knowFile string,
	timeout time.Duration,
	knownHostCheck, askAddKnownHost bool,
) (*Client, error) {

	config, err := createConfig(user, addr, port, passwd, privateKey, privateKeyRaw, timeout)
	if err != nil {
		return nil, err
	}

	if knownHostCheck {
		config.Callback = VerifyHost(knowFile, askAddKnownHost)
	} else {
		config.Callback = ssh.InsecureIgnoreHostKey()
	}

	c, err := NewConn(config)

	return c, err
}

// NewConn returns a new client and error if any.
func NewConn(config *Config) (c *Client, err error) {
	c = &Client{
		Config: config,
	}

	c.Client, err = Dial("tcp", config)
	return
}

// Dial starts a client connection to SSH server based on config.
func Dial(proto string, c *Config) (*ssh.Client, error) {
	return ssh.Dial(proto, net.JoinHostPort(c.Addr, fmt.Sprint(c.Port)), &ssh.ClientConfig{
		User:            c.User,
		Auth:            c.Auth,
		Timeout:         c.Timeout,
		HostKeyCallback: c.Callback,
		BannerCallback:  c.BannerCallback,
	})
}

func (c *Client) session() (*ssh.Session, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Client == nil {
		return nil, errors.New("ssh client not initialized")
	}

	session, err := c.Client.NewSession()
	if err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	err = session.RequestPty("xterm", 100, 50, modes)
	if err != nil {
		return nil, err
	}

	log.Infof("failed to set LANG to en_US.UTF-8. (Error: %v)", err)
	log.Errorf("failed to set LANG to en_US.UTF-8. (Error: %v)", errors.Errorf("123455"))

	if err := session.Setenv("LANG", "en_US.UTF-8"); err != nil {
		log.Infof("failed to set LANG to en_US.UTF-8. (Error: %v)", err)
		log.Errorf("failed to set LANG to en_US.UTF-8. (Error: %v)", err)
	}

	return session, nil
}

func (c *Client) Run(cmd string) (stdout []byte, err error) {
	session, err := c.session()
	if err != nil {
		return nil, err
	}
	defer func(session *ssh.Session) {
		dErr := session.Close()
		if dErr != nil && dErr != io.EOF && err == nil {
			err = dErr
		}
	}(session)

	//err = session.Start(strings.TrimSpace(cmd))
	//if err != nil {
	//	return nil, err
	//}
	//
	//err = session.Wait()
	//if err != nil {
	//	fmt.Println(err)
	//}

	return session.CombinedOutput(strings.TrimSpace(cmd))
}

func createConfig(
	user,
	addr string,
	port uint,
	passwd,
	privateKey,
	privateKeyRaw string,
	timeout time.Duration,
) (*Config, error) {
	c := &Config{}
	if len(user) == 0 {
		return nil, errors.New("username is required for ssh connection")
	}

	if len(addr) == 0 {
		return nil, errors.New("address is required for ssh connection")
	}

	if !iputils.IsValidIP(addr) && !iputils.IsValidDomain(addr) {
		return nil, errors.Errorf("address is an invalid ip or domain address: %s", addr)
	}

	if len(passwd) == 0 && len(privateKey) == 0 && len(privateKeyRaw) == 0 {
		return nil, errors.New("private key or private key is required")
	}

	c.User = user

	c.Addr = addr

	c.Port = setSSHPort(port)

	if timeout == 0 {
		c.Timeout = DefaultTimeout
	}

	auth := Auth{}

	if len(passwd) > 0 {
		auth = append(auth, Password(passwd))
	}

	if len(privateKey) > 0 {
		keyAuth, err := Key(privateKey, "")
		if err != nil {
			return nil, errors.Wrap(err, "private key parse failed")
		}
		auth = append(auth, keyAuth)
	}

	if len(privateKey) == 0 && len(privateKeyRaw) > 0 {
		keyAuth, err := RawKey(privateKey, "")
		if err != nil {
			return nil, errors.Wrap(err, "private key parse failed")
		}
		auth = append(auth, keyAuth)
	}

	c.Auth = auth

	return c, nil
}

func setSSHPort(port uint) uint {
	var p uint
	if port > 0 && port < 65535 {
		p = port
	} else {
		p = DefaultSSHPort
	}
	return p
}
