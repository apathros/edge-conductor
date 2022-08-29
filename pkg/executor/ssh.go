/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package executor

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	host     string
	user     string
	password string
	key      string
	port     int
	config   *ssh.ClientConfig
	client   *ssh.Client
}

func (c *sshClient) Connect() error {
	c.config = &ssh.ClientConfig{
		Config: ssh.Config{
			Ciphers: []string{
				"aes256-gcm@openssh.com",
				"chacha20-poly1305@openssh.com",
				"aes256-ctr", "aes256-cbc"},
		},
		Timeout: time.Second * 5,
		User:    c.user,
		// #nosec G106
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(c.password)},
	}

	if c.key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(c.key))
		if err == nil {
			c.config.Auth = append(c.config.Auth, ssh.PublicKeys(signer))
		}
	}

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	client, err := ssh.Dial("tcp", addr, c.config)
	c.client = client
	return err
}

func (c *sshClient) Disconnect() error {
	return c.client.Close()
}

func (c *sshClient) CmdWithAttachIO(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	cmdStr := strings.Join(cmd, " ")
	log.Debugf("CmdWithAttachIO: cmd: %v", cmdStr)
	session, err := c.client.NewSession()
	if err != nil {
		return err
	}
	done := make(chan struct{})
	defer func() {
		done <- struct{}{}
	}()
	go func() {
		select {
		case <-ctx.Done():
			log.Infof("cancel cmd execution")
		case <-done:
			log.Debugf("cmd done")
		}
		/* From the issue description, always getting EOF(End Of File)
		 * when session.Close after session.Run,
		 * https://github.com/golang/go/issues/38115
		 * here ignore EOF, doesn't think it as an error case
		 */
		if err := session.Close(); err != nil && err != io.EOF {
			log.Errorf("Failed to close session.")
		}
	}()
	if tty {
		modes := ssh.TerminalModes{
			ssh.ECHO: 0,
		}
		if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
			return err
		}
	}
	if stdin != nil {
		session.Stdin = stdin
	}
	if stdout != nil {
		session.Stdout = stdout
	}
	if stderr != nil {
		session.Stderr = stderr
	}
	return session.Run(cmdStr)
}
