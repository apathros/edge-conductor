/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capiprovisionbinarydownload

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	capiutils "github.com/intel/edge-conductor/pkg/eputils/capiutils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

const (
	REGSERVERCERTFILE = "cert/pki/registry/registry.pem"
)

func launchIpaDownload(workFolder string, clusterConfig *pluginapi.CapiClusterConfig, tmpl *capiutils.CapiTemplate) error {
	var err error

	dstFile := filepath.Join(workFolder, "ironic-containers.yaml")
	err = capiutils.TmplFileRendering(tmpl, workFolder, clusterConfig.BaremetelOperator.IronicContainers, dstFile)
	if err != nil {
		log.Errorf("Failed to render %s, %v", clusterConfig.BaremetelOperator.IronicContainers, err)
		return err
	}

	var ironicContainers pluginapi.Containers
	err = eputils.LoadSchemaStructFromYamlFile(&ironicContainers, dstFile)
	if err != nil {
		log.Errorf("Load capi cluster config failed, %v", err)
		return err
	}

	for _, c := range ironicContainers.Containers {
		if c.Name == "ipa-downloader" {
			err = docker.DockerRun(c)
			if err != nil {
				log.Errorf("Container %s run fail, %v", c.Name, err)
				return err
			}
		}
	}

	return nil
}

func copyIronicOsPrivisionImage(ep_params *pluginapi.EpParams, ironicHttpdFolder string, capiSetting *pluginapi.CapiSetting) error {
	var err error

	if capiSetting.IronicConfig.IronicOsImage != "" {
		osFileName := filepath.Base(capiSetting.IronicConfig.IronicOsImage)
		provisionOsImage := filepath.Join(ironicHttpdFolder, osFileName)

		osImagePath := filepath.Join(ep_params.Workspace, osFileName)

		if !eputils.FileExists(osImagePath) {
			log.Warnf("No OS image in Workspace, %s", osImagePath)
			return nil
		}

		if !eputils.FileExists(provisionOsImage) {
			log.Infof("Copy OS image for provision, %s", osImagePath)
			log.Infof("Please wait for a while")

			_, err = eputils.CopyFile(provisionOsImage, osImagePath)
			if err != nil {
				return err
			}

			sha256sumFilePath := provisionOsImage + ".shasum"
			sha256sum, _ := eputils.GenFileSHA256(provisionOsImage)
			err = eputils.WriteStringToFile(sha256sum, sha256sumFilePath)
			if err != nil {
				log.Errorf("Failed to write data %s, reason: %v", sha256sumFilePath, err)
				return err
			}
		} else {
			log.Infof("Os image is already ready")
		}
	}

	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_cluster_manifest := input_cluster_manifest(in)

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
		return eputils.GetError("errProvider")
	} else {
		provider = providers[0]
	}

	workFolder := filepath.Join(input_ep_params.Runtimedir, provider)
	if err = eputils.CreateFolderIfNotExist(workFolder); err != nil {
		return err
	}

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

	if provider == capiutils.CAPI_METAL3 {
		ironic_data_dir := filepath.Join(workFolder, "ironic")
		ironic_image_dir := filepath.Join(ironic_data_dir, "html", "images")
		err = os.MkdirAll(ironic_data_dir, 0755)
		if err != nil {
			return err
		}

		err = os.MkdirAll(ironic_image_dir, 0755)
		if err != nil {
			return err
		}

		err = launchIpaDownload(workFolder, &clusterConfig, &tmpl)
		if err != nil {
			log.Errorf("Ironic provision agent download failed, %v", err)
			return err
		}

		err = copyIronicOsPrivisionImage(input_ep_params, ironic_image_dir, &capiSetting)
		if err != nil {
			log.Errorf("Copy ironic os provision image fail, %v", err)
			return err
		}
	}

	return nil
}
