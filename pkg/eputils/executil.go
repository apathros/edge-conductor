/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package eputils

//go:generate mockgen -destination=./mock/executil_mock.go -package=mock -copyright_file=../../api/schemas/license-header.txt ep/pkg/eputils ExecWrapper

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	log "github.com/sirupsen/logrus"
)

var (
	errGeneral = fmt.Errorf("")
)

type ExecWrapper interface {
	RunCMD(cmd *exec.Cmd) (string, error)
	RunCMDEx(cmd *exec.Cmd, useOsOut bool) (string, error)
}

func RunCMD(cmd *exec.Cmd) (string, error) {
	return RunCMDEx(cmd, false)
}

func RunCMDEx(cmd *exec.Cmd, useOsOut bool) (string, error) {
	var stdout, stderr bytes.Buffer
	if useOsOut {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	}
	err := cmd.Run()

	if !useOsOut {
		outStr, errStr := stdout.String(), stderr.String()
		if err != nil {
			if len(outStr) > 0 {
				log.Infof("out:\n%s\n", outStr)
			}
			if len(errStr) > 0 {
				log.Errorf("err:\n%s\n", errStr)
			}
			log.Infof("%s\n%s\n%s\n", err, outStr, errStr)
			return "", errGeneral
		}
		return outStr, nil
	}
	return "", err
}
