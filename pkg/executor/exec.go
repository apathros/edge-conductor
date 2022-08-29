/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package executor

import (
	"context"
	"io"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

type day0Client struct {
}

func (c *day0Client) Connect() error {
	return nil
}

func (c *day0Client) Disconnect() error {
	return nil
}

func (c *day0Client) CmdWithAttachIO(ctx context.Context, cmd []string, stdin io.Reader, stdout, stderr io.Writer, tty bool) error {
	log.Debugf("CmdWithAttachIO: cmd: %v", cmd)
	execCmd := exec.CommandContext(ctx, cmd[0], cmd[1:]...)
	if stdin != nil {
		execCmd.Stdin = stdin
	}
	if stdout != nil {
		execCmd.Stdout = stdout
	}
	if stderr != nil {
		execCmd.Stderr = stderr
	}
	return execCmd.Run()
}
