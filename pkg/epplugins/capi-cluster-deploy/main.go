/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capiclusterdeploy

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	capiutils "github.com/intel/edge-conductor/pkg/eputils/capiutils"
	kubeutils "github.com/intel/edge-conductor/pkg/eputils/kubeutils"
	serviceutil "github.com/intel/edge-conductor/pkg/eputils/service"

	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	TIMEOUT     = 360
	WAIT_10_SEC = 10
	WAIT_1_SEC  = 1
)

func applyCluster(ep_params *pluginapi.EpParams, workFolder, mClusterConfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	log.Debugf("ep_params: %v", ep_params)

	var err error

	dstFile := filepath.Join(workFolder, "cluster.yaml")
	err = capiutils.TmplFileRendering(tmpl, workFolder, clusterConfig.WorkloadCluster.URL, dstFile)
	if err != nil {
		log.Errorf("Failed to render %s, %v", clusterConfig.BaremetelOperator.URL, err)
		return err
	}

	name := clusterConfig.WorkloadCluster.Name
	deployer := serviceutil.NewYamlDeployer(name, clusterConfig.WorkloadCluster.Namespace, dstFile)
	err = deployer.YamlInstall(mClusterConfig)
	defer func() {
		err := os.RemoveAll(dstFile)
		if err != nil {
			log.Errorf("Fail to remove file, %v", err)
		}
	}()

	if err != nil {
		log.Errorf("Apply %s fail, %v", clusterConfig.WorkloadCluster.URL, err)
		return err
	}

	return nil
}

func genClusterKubeconfig(ep_params *pluginapi.EpParams, workFolder, mClusterConfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate, kubeconfig *pluginapi.Filecontent) error {
	log.Debugf("ep_params: %v", ep_params)
	log.Debugf("workFolder: %v", workFolder)
	log.Debugf("tmpl: %v", tmpl)
	var err error

	secretName := clusterConfig.WorkloadCluster.Name + "-kubeconfig"
	kubeconfigSecret, err := kubeutils.NewSecret(clusterConfig.WorkloadCluster.Namespace, secretName, "", mClusterConfig)
	if err != nil {
		log.Errorf("Failed to get secret in namespace %s", clusterConfig.WorkloadCluster.Namespace)
		return err
	}

	err = kubeconfigSecret.Get()
	if err != nil {
		return err
	}

	byteData := kubeconfigSecret.GetData()
	kubeconfig.Content = string(byteData["value"])
	return nil
}

func checkProvisionedMachine(ep_params *pluginapi.EpParams, workFolder, mClusterConfig string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	log.Debugf("workFolder: %v", workFolder)
	log.Debugf("tmpl: %v", tmpl)
	ready := false
	count := 0

	for count < TIMEOUT {
		cmd := exec.Command(ep_params.Workspace+"/kubectl", "get", "machine", "-n", clusterConfig.WorkloadCluster.Namespace, "--kubeconfig", mClusterConfig)
		outputStr, err := eputils.RunCMD(cmd)
		if err != nil {
			log.Errorf("Failed to get workload config. %v", err)
			return err
		}

		reAvaliable := regexp.MustCompile(`.*\sRunning\s.*`)
		matchesOutput := reAvaliable.FindAllStringSubmatch(outputStr, -1)

		if len(matchesOutput) < 1 {
			log.Infof("sleep %d sec", WAIT_10_SEC)
			time.Sleep(WAIT_10_SEC * time.Second)
			count++
		} else {
			ready = true
			break
		}
	}

	if !ready {
		log.Errorf("No controlplane node is not ready, please check")
		return eputils.GetError("errNode")
	}

	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_cluster_manifest := input_cluster_manifest(in)
	output_kubeconfig := output_kubeconfig(outp)

	var err error
	var provider string
	providers := make([]string, 0)
	for _, p := range input_ep_params.Kitconfig.Parameters.Extensions {
		for _, i := range capiutils.InfraProviderList {
			if p == i {
				providers = append(providers, p)
			}
		}
	}

	if len(providers) != 1 {
		log.Errorf("Please select one provider")
		return eputils.GetError("errProvider")
	} else {
		provider = providers[0]
	}

	workFolder := filepath.Join(input_ep_params.Runtimedir, provider)
	var clusterConfig pluginapi.CapiClusterConfig
	clusterConfig.WorkloadCluster = new(pluginapi.CapiClusterConfigWorkloadCluster)
	clusterConfig.BaremetelOperator = new(pluginapi.CapiClusterConfigBaremetelOperator)
	err = eputils.LoadSchemaStructFromYamlFile(&clusterConfig, input_ep_params.Kitconfig.Cluster.Config)
	if err != nil {
		log.Errorf("Load capi cluster config failed, %v", err)
		return err
	}

	var capiSetting pluginapi.CapiSetting
	capiSetting.Provider = provider
	capiSetting.InfraProvider = new(pluginapi.CapiSettingInfraProvider)
	capiSetting.IronicConfig = new(pluginapi.CapiSettingIronicConfig)

	var tmpl capiutils.CapiTemplate
	err = capiutils.GetCapiSetting(input_ep_params, input_cluster_manifest, &clusterConfig, &capiSetting)
	if err != nil {
		log.Errorf("CapiHostProvision, get CapiSetting failed, %v", err)
		return err
	}
	err = capiutils.GetCapiTemplate(input_ep_params, capiSetting, &tmpl)
	if err != nil {
		log.Errorf("CapiHostProvision, get CapiTemplate failed, %v", err)
		return err
	}

	mClusterConfig := capiutils.GetManagementClusterKubeconfig(input_ep_params)
	err = applyCluster(input_ep_params, workFolder, mClusterConfig, &clusterConfig, &tmpl)
	if err != nil {
		log.Errorf("Cluster %s apply fail, %v", clusterConfig.WorkloadCluster.Name, err)
		return err
	}

	err = checkProvisionedMachine(input_ep_params, workFolder, mClusterConfig, &clusterConfig, &tmpl)
	if err != nil {
		log.Errorf("Failed to provisioned control plane, %v", err)
		return err
	}

	err = genClusterKubeconfig(input_ep_params, workFolder, mClusterConfig, &clusterConfig, &tmpl, output_kubeconfig)
	if err != nil {
		log.Errorf("Failed to get cluster %s kubeconfig, %v", clusterConfig.WorkloadCluster.Name, err)
		return err
	}

	return nil
}
