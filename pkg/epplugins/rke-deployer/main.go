/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package rkedeployer

import (
	eputils "ep/pkg/eputils"
	repoutils "ep/pkg/eputils/repoutils"
	"ep/pkg/executor"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_eptopcfg := input_ep_params.Kitconfig
	input_files := input_files(in)
	output_kubeconfig := output_kubeconfig(outp)

	rkeCfgSrc := ""
	rkeCfgDir := ""
	if input_eptopcfg != nil && input_eptopcfg.Cluster != nil {
		rkeCfgSrc = input_eptopcfg.Cluster.Config
		rkeCfgDir = input_eptopcfg.Cluster.ExportConfigFolder
	}
	log.Infof("Read cluster config file from %v", rkeCfgSrc)
	rkeCfgContent, err := eputils.LoadJsonFile(rkeCfgSrc)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	if rkeCfgDir == "" {
		if home, err := os.UserHomeDir(); err != nil {
			return err
		} else {
			rkeCfgDir = filepath.Join(home, ".ec", "rke", "cluster")
		}
	}
	err = eputils.MakeDir(rkeCfgDir)
	if err != nil {
		return err
	}

	if len(input_files.Files) == 0 {
		return eputils.GetError("errInputArryEmpty")
	}
	log.Debugf("inputfiles:%v", input_files.Files[0])
	rkeCfgTgt := filepath.Join(rkeCfgDir, "rke_cluster.yml")
	err = eputils.WriteStringToFile(string(rkeCfgContent), rkeCfgTgt)
	if err != nil {
		return err
	}

	rkeBin := filepath.Join(input_ep_params.Runtimebin, "rke")

	err = repoutils.PullFileFromRepo(rkeBin, input_files.Files[0].Mirrorurl)
	if err != nil {
		log.Errorf("%s", err)
		return eputils.GetError("errPullingFile")
	}

	err = os.Chmod(rkeBin, 0700)
	if err != nil {
		return err
	}

	err = executor.Run("config/executor/rke_preflight.yml", input_ep_params, nil)
	if err != nil {
		return err
	}

	log.Debugf("Cluster.Config: %s", rkeCfgTgt)

	var cmd *exec.Cmd
	if log.DebugLevel == log.GetLevel() {
		cmd = exec.Command(rkeBin, "-d", "up", "--config", rkeCfgTgt)
	} else {
		cmd = exec.Command(rkeBin, "up", "--config", rkeCfgTgt)
	}

	log.Infof("Deploying rke...")
	_, err = eputils.RunCMDEx(cmd, true)
	if err != nil {
		log.Errorf("Failed to create RKE cluster. %s", err)
		return err
	}

	content, err := eputils.LoadJsonFile(filepath.Join(rkeCfgDir, "kube_config_rke_cluster.yml"))
	if err != nil {
		return err
	}
	output_kubeconfig.Content = string(content)
	log.Debugf("%v", output_kubeconfig)

	return nil
}
