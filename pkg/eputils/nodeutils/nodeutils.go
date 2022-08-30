/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package nodeutils

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
)

func FindNodeInClusterByIP(nodelist *corev1.NodeList, ip string) bool {
	if nodelist == nil {
		return false
	}
	for _, nodeinfo := range nodelist.Items {
		for _, Addresses := range nodeinfo.Status.Addresses {
			ipaddress := Addresses.Address
			if len(ipaddress) != 0 && ip == ipaddress {
				log.Infof("Node %s is in current node list ", ip)
				return true
			}
		}
	}
	return false
}

func GetClusterVersion(nodeList *corev1.NodeList) string {
	if nodeList == nil {
		return ""
	}
	for _, nodeinfo := range nodeList.Items {
		return nodeinfo.Status.NodeInfo.KubeletVersion
	}
	return ""
}

func GetCRI(nodeList *corev1.NodeList) string {
	if nodeList == nil {
		return ""
	}
	for _, nodeinfo := range nodeList.Items {
		return nodeinfo.Status.NodeInfo.ContainerRuntimeVersion
	}
	return ""
}

func GetKubeConfigContent(kubeConfigFile string) (*pluginapi.Filecontent, error) {
	configcontent, err := ioutil.ReadFile(kubeConfigFile)
	if err != nil {
		log.Errorf("Failed to read file %s: %s", kubeConfigFile, err)
		return nil, err
	}

	output_kubeconfig := &pluginapi.Filecontent{
		Content: string(configcontent),
	}
	return output_kubeconfig, nil
}
