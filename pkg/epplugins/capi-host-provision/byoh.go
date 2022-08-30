/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
package capihostprovision

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	capiutils "github.com/intel/edge-conductor/pkg/eputils/capiutils"
	kubeutils "github.com/intel/edge-conductor/pkg/eputils/kubeutils"
	"github.com/intel/edge-conductor/pkg/executor"
	"os/exec"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	REGSERVERCERTFILE = "cert/pki/registry/registry.pem"
)

func DeploymentReady(management_kubeconfig, namespace, deploymentName string) error {
	count := 0

	MgrDeployment, err := kubeutils.NewDeployment(namespace, deploymentName, "", management_kubeconfig)
	if err != nil {
		log.Errorf("Failed to new deployment object %v", err)
		return err
	}

	for count < TIMEOUT {
		err = MgrDeployment.Get()
		if err != nil {
			log.Errorf("Failed to get deployment %s, %v", deploymentName, err)
			return err

		}

		status := MgrDeployment.GetStatus()
		if status.ReadyReplicas == status.Replicas {
			log.Infof("Deployment %s: is ready", deploymentName)
			break
		}

		log.Infof("Deployment %s is not ready, waiting", deploymentName)
		time.Sleep(WAIT_10_SEC * time.Second)
		count++
	}

	if count >= TIMEOUT {
		log.Errorf("Deployment %s: launch fail", deploymentName)
		return eputils.GetError("errDeploymentLaunchFail")
	}

	return nil
}

func checkByoHosts(ep_params *pluginapi.EpParams, workFolder, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	ready := false
	count := 0

	for count < TIMEOUT {
		cmd := exec.Command(ep_params.Workspace+"/kubectl", "get", "byohosts", "-n", clusterConfig.WorkloadCluster.Namespace, "--kubeconfig", management_kubeconfig)
		outputStr, err := eputils.RunCMD(cmd)
		if err != nil {
			log.Errorf("Failed to get workload config. %v", err)
			return err
		}

		lines := strings.Count(outputStr, "\n")
		if lines < 2 {
			log.Infof("sleep %d sec", WAIT_10_SEC)
			time.Sleep(WAIT_10_SEC * time.Second)
			count++
		} else {
			ready = true
			break
		}
	}

	if !ready {
		log.Errorf("Node is not ready, please check")
		return eputils.GetError("errNodeNotReady")
	}

	return nil
}

func byohHostProvision(ep_params *pluginapi.EpParams, workFolder, management_kubeconfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	var err error

	err = DeploymentReady(management_kubeconfig, "byoh-system", "byoh-controller-manager")
	if err != nil {
		log.Errorf("ByohCtlMgr deployment launch fail, %v", err)
		return err
	}

	err = executor.Run(clusterConfig.ByohAgent.InitScript, ep_params, tmpl.CapiSetting)
	if err != nil {
		log.Errorf("ByohAgent pre-provision failed, %v", err)
		return err
	}

	err = checkByoHosts(ep_params, workFolder, management_kubeconfig, clusterConfig, tmpl)
	if err != nil {
		log.Errorf("Failed to get available byohost, %v", err)
		return err
	}

	return nil
}
