/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package nodejoinprepare

import (
	// TODO: Add Plugin Imports Here
	"encoding/base64"
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	kubeutils "ep/pkg/eputils/kubeutils"
	nodeutils "ep/pkg/eputils/nodeutils"
	"ep/pkg/executor"
	"fmt"
	log "github.com/sirupsen/logrus"
	"path"
	"strings"
)

type NodeJoinInfo struct {
	Version      string
	CRI          CRI
	RegistryAuth string
	Provider     string
}

type CRI struct {
	Name    string
	Version string
}

const (
	ORAS_BINARY_URL             = "https://github.com/oras-project/oras/releases/download/v0.13.0/oras_0.13.0_linux_amd64.tar.gz"
	NODE_JOIN_PREPARE_FILE_PATH = "config/executor/node-join-prepare.yml"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	runtime_kubeconfig := input_ep_params.Kubeconfig

	log.Infof("Plugin: node-join-prepare")

	var newNodeList []*pluginapi.Node

	input_kubeconfig, err := nodeutils.GetKubeConfigContent(runtime_kubeconfig)
	if err != nil {
		log.Errorf("get kube config content failed: %v", err)
		return err
	}

	var provider string
	for _, extension := range input_ep_params.Extensions {
		if strings.Contains(extension.Name, "capi") {
			provider = strings.TrimPrefix(extension.Name, "capi-")
		}
	}

	targetFile := fmt.Sprintf("%s/capi-%s/%s", input_ep_params.Runtimedir, provider, path.Base(ORAS_BINARY_URL))

	if err = eputils.DownloadFile(targetFile, ORAS_BINARY_URL); err != nil {
		log.Errorf("download oras tool failed: %v", err)
		return err
	}

	nodelist, err := kubeutils.GetNodeList(input_kubeconfig, "")
	if err != nil {
		log.Errorf("Failed to get node list. %s", err)
		return err
	}
	version := nodeutils.GetClusterVersion(nodelist)
	cri := nodeutils.GetCRI(nodelist)
	log.Infof("cluster version is %v, cluster runtime is %v", version, cri)

	AuthStr := base64.StdEncoding.EncodeToString([]byte(input_ep_params.Kitconfig.Parameters.Customconfig.Registry.User + ":" + input_ep_params.Kitconfig.Parameters.Customconfig.Registry.Password))

	nodeJoinInfo := NodeJoinInfo{
		Version: version,
		CRI: CRI{
			Name:    strings.Split(cri, "://")[0],
			Version: strings.Split(cri, "://")[1],
		},
		RegistryAuth: AuthStr,
		Provider:     provider,
	}

	for _, node := range input_ep_params.Kitconfig.Parameters.Nodes {
		if node.IP == "" || nodeutils.FindNodeInClusterByIP(nodelist, node.IP) {
			log.Infof("Node(%s) already joined in the cluster!", node.IP)
			continue
		}
		newNodeList = append(newNodeList, node)
	}

	input_ep_params.Kitconfig.Parameters.Nodes = newNodeList

	err = executor.Run(fmt.Sprintf("%s/%s", input_ep_params.Workspace, NODE_JOIN_PREPARE_FILE_PATH), input_ep_params, nodeJoinInfo)
	if err != nil {
		log.Errorf("ByohAgent pre-provision failed, %v", err)
		return err
	}

	return nil
}
