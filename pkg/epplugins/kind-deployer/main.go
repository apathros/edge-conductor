/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package kinddeployer

import (
	"fmt"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	repoutils "github.com/intel/edge-conductor/pkg/eputils/repoutils"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_files := input_files(in)
	input_kind_config := input_kind_config(in)
	output_kubeconfig := output_kubeconfig(outp)

	kubeconfig_dir := filepath.Join(input_ep_params.Runtimedir, ".kube")
	kubeconfig := filepath.Join(kubeconfig_dir, "config")

	err := os.MkdirAll(kubeconfig_dir, 0700)
	if err != nil {
		return err
	}

	kindbin := filepath.Join(input_ep_params.Runtimebin, "kind")
	if len(input_files.Files) > 0 {
		err = repoutils.PullFileFromRepo(kindbin, input_files.Files[0].Mirrorurl)
		if err != nil {
			return err
		}
	} else {
		log.Error("No valid KIND binary found. Try to run \"cluster build\" first.")
		return eputils.GetError("errCreateKIND")
	}
	err = os.Chmod(kindbin, 0700)
	if err != nil {
		return err
	}
	kindClusterInstanceCfgTgt := filepath.Join(input_ep_params.Runtimedata, "kind_cluster_instance_cfg.yml")
	if _, err := os.Stat(input_ep_params.Runtimedata); os.IsNotExist(err) {
		if err = os.MkdirAll(input_ep_params.Runtimedata, os.ModePerm); err != nil {
			log.Error(err)
			return err
		}
	}
	if err := ioutil.WriteFile(kindClusterInstanceCfgTgt, []byte(input_kind_config.Content), 0600); err != nil {
		log.Errorf("Fail to write file %s", kindClusterInstanceCfgTgt)
		return err
	}
	cmd := exec.Command(kindbin, "create", "cluster", "--config", kindClusterInstanceCfgTgt)
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("KUBECONFIG=%s", kubeconfig),
	)
	log.Infof("Deploying kind...")
	_, err = eputils.RunCMDEx(cmd, true)
	if err != nil {
		log.Errorf("Failed to create KIND cluster. %s", err)
		return err
	}

	err = eputils.RemoveFile(kindClusterInstanceCfgTgt)
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(kubeconfig)
	if err != nil {
		return err
	}
	output_kubeconfig.Content = string(content)

	return nil
}
