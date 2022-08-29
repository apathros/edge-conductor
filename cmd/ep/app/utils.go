/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package app

import (
	epapiplugins "ep/pkg/api/plugins"
	"fmt"
	"net"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

var (
	epPath         = ""
	kitTopConfig   = ""
	runtimePath    = ""
	runtimeDataDir = ""
)

func GetWorkspacePath() string {
	epPath = os.Getenv("EPPATH")
	if len(epPath) == 0 {
		_, err := os.Readlink(os.Args[0])
		if err != nil {
			v, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			epPath = fmt.Sprintf("%s/..", v)
			return epPath
		}
		v, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		return v
	}
	return epPath
}

func ChangeToWorkspacePath() error {
	epPath = os.Getenv("EPPATH")
	if len(epPath) == 0 {
		_, err := os.Readlink(os.Args[0])
		if err != nil {
			v, _ := filepath.Abs(filepath.Dir(os.Args[0]))
			epPath = fmt.Sprintf("%s/..", v)
			err := os.Chdir(epPath)
			return err
		}
		return nil
	}
	return nil
}
func MakeDir(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return err
	}
	return nil
}

func Utils_Init() error {
	epPath := GetWorkspacePath()
	err := ChangeToWorkspacePath()
	if err != nil {
		return err
	}
	kitTopConfig = fnTopConfig
	runtimePath = filepath.Join(epPath, dirRuntime)
	runtimeDataDir = filepath.Join(runtimePath, fnRuntimeDataDir)
	return nil
}

func GetRuntimeFolder() string {
	return runtimePath
}

func FileNameofRuntime(target_name string) (string, error) {
	runtimeFolder := GetRuntimeFolder()
	_, err := os.Stat(runtimeFolder)
	if err != nil {
		err = MakeDir(runtimeFolder)
		if err != nil {
			return "", err
		}
	}
	targetfile := filepath.Join(runtimeFolder, target_name)
	return targetfile, nil
}

func GetDefaultTopConfigName() string {
	return kitTopConfig
}

func GetRuntimeTopConfig(epParams *epapiplugins.EpParams) *epapiplugins.Kitconfig {
	return epParams.Kitconfig
}

func GetHostDefaultIP() string {
	// Fake udp connection to get default local ip
	conn, _ := net.Dial("udp", "8.8.8.8:80")
	defer conn.Close()
	defaultAddr := conn.LocalAddr().(*net.UDPAddr)

	return defaultAddr.IP.String()
}

func GetDefaultKubeConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorln("Failed to get HOME dir:", err)
		return ""
	}
	return fmt.Sprintf("%s/.kube/config", home)
}
