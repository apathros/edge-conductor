/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package nodejoindeploy

import (
	// TODO: Add Plugin Imports Here
	"bytes"
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	kubeutils "ep/pkg/eputils/kubeutils"
	nodeutils "ep/pkg/eputils/nodeutils"
	"fmt"
	log "github.com/sirupsen/logrus"
	kubeadmcmd "k8s.io/kubernetes/cmd/kubeadm/app/cmd"
	cmdutil "k8s.io/kubernetes/cmd/kubeadm/app/cmd/util"
	"strings"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	runtime_kubeconfig := input_ep_params.Kubeconfig

	// TODO: Add Plugin Code Here
	log.Infof("Plugin: node-join")
	log.Infof("%v", input_ep_params)

	input_kubeconfig, err := nodeutils.GetKubeConfigContent(runtime_kubeconfig)
	if err != nil {
		log.Errorf("get kube config content failed: %v", err)
		return err
	}

	nodelist, err := kubeutils.GetNodeList(input_kubeconfig, "")
	if err != nil {
		log.Errorf("Failed to get node list. %s", err)
		return err
	}

	var cmd string
	cri := nodeutils.GetCRI(nodelist)

	for _, node := range input_ep_params.Kitconfig.Parameters.Nodes {
		if node.IP == "" || nodeutils.FindNodeInClusterByIP(nodelist, node.IP) {
			log.Infof("Node(%s) already joined in the cluster!", node.IP)
			continue
		}

		joinCMD, err := GetNodeJoinCMD(input_kubeconfig, runtime_kubeconfig)
		if err != nil {
			log.Errorf("get kubeadm token failed: %v", err)
		}

		nodeAddr := fmt.Sprintf("%s:%d", node.IP, node.SSHPort)
		sshcfg, err := eputils.GenSSHConfig(node)
		if err != nil {
			log.Errorf("Fail to gen config %v", err)
			return err
		}

		if strings.Contains(cri, "containerd") {
			cmd = fmt.Sprintf("sudo %s --cri-socket=unix:///run/containerd/containerd.sock", joinCMD)
		} else {
			cmd = fmt.Sprintf("sudo %s", joinCMD)
		}

		err = eputils.RunRemoteCMD(nodeAddr, sshcfg, cmd)
		if err != nil {
			log.Errorf("Failed to enable containerd %v ", err)
			return err
		}
	}

	return nil
}

func GetNodeJoinCMD(input_kubeconfig *pluginapi.Filecontent, kubeConfig string) (string, error) {
	buf := new(bytes.Buffer)
	client, err := kubeutils.ClientFromEPKubeConfig(input_kubeconfig)
	if err != nil {
		log.Errorf("Failed to Get ClientSet. %s", err)
		return "", err
	}
	cfg := cmdutil.DefaultInitConfiguration()
	err = kubeadmcmd.RunCreateToken(buf, client, "", cfg, true, "", kubeConfig)
	if err != nil {
		log.Errorln(err)
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}
