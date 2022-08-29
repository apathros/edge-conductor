/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Template auto-generated once, maintained by plugin owner.

package capiparser

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	capiutils "ep/pkg/eputils/capiutils"
	cutils "ep/pkg/eputils/conductorutils"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func getManagementClusterBinaryList(kindprovider *pluginapi.ClustermanifestClusterProvidersItems0, images *pluginapi.Images, files *pluginapi.Files) error {
	kindImage, err := cutils.GetImageFromProvider(kindprovider, "img_node")
	if err != nil {
		log.Errorln("Failed to find container image for KIND node.")
		return err
	}
	kindHAProxy, err := cutils.GetImageFromProvider(kindprovider, "img_haproxy")
	if err != nil {
		log.Errorln("Failed to find container image for KIND haproxy.")
		return err
	}
	kindBin, kindSHA256, err := cutils.GetBinaryFromProvider(kindprovider, "kindtool")
	if err != nil {
		log.Errorln("Failed to find binary for KIND.")
		return err
	}

	images.Images = append(images.Images, &pluginapi.ImagesItems0{Name: "kind", URL: kindImage})
	images.Images = append(images.Images, &pluginapi.ImagesItems0{Name: "kindhaproxy", URL: kindHAProxy})

	files.Files = append(files.Files, &pluginapi.FilesItems0{
		URL:      kindBin,
		Hash:     kindSHA256,
		Hashtype: "sha256",
		Urlreplacement: &pluginapi.FilesItems0Urlreplacement{
			New:    "capi/kind",
			Origin: eputils.GetBaseUrl(kindBin),
		},
	})

	if files.Files == nil {
		return eputils.GetError("errAppendFile")
	}

	return nil
}

func getDockerImagesList(manifest []*pluginapi.ClustermanifestCapiClusterProvidersItems0, infra_provider capiutils.CapiInfraProvider, images *pluginapi.Images) error {
	capi_cluster_name := capiutils.GetManifestConfigNameByCapiInfraProvider(infra_provider)

	images.Images = []*pluginapi.ImagesItems0{}

	clusterProviderItem, err := capiutils.GetCapiClusterProviderConfig(manifest, capi_cluster_name)
	if err != nil {
		return err
	}

	for _, image := range clusterProviderItem.Images {
		images.Images = append(images.Images, &pluginapi.ImagesItems0{Name: "", URL: image})
	}

	return nil
}

func generateFileItemsByURL(url string, subpath string) *pluginapi.FilesItems0 {
	dir := eputils.GetBaseUrl(url)
	subRef := filepath.Join("capi", subpath)
	return &pluginapi.FilesItems0{
		URL: url,
		Urlreplacement: &pluginapi.FilesItems0Urlreplacement{
			New:    subRef,
			Origin: dir,
		},
	}
}

func getProviderBinaryList(manifest []*pluginapi.ClustermanifestCapiClusterProvidersItems0, infra_provider capiutils.CapiInfraProvider, files *pluginapi.Files) error {
	capi_cluster_name := capiutils.GetManifestConfigNameByCapiInfraProvider(infra_provider)

	files.Files = []*pluginapi.FilesItems0{}

	clusterProviderItem, err := capiutils.GetCapiClusterProviderConfig(manifest, capi_cluster_name)
	if err != nil {
		return err
	}

	if clusterProviderItem.Binaries == nil {
		return eputils.GetError("errBinaries")
	}
	for _, binaries := range clusterProviderItem.Binaries {
		files.Files = append(files.Files, generateFileItemsByURL(binaries.URL, binaries.Name))
	}

	for _, provider := range clusterProviderItem.Providers {
		if provider == nil {
			log.Warnln("Invalid parameter: nil provider in manifest.")
			continue
		}
		subpath := filepath.Join(provider.Parameters.ProviderLabel, provider.Parameters.Version)
		files.Files = append(files.Files, generateFileItemsByURL(provider.URL, subpath))
		files.Files = append(files.Files, generateFileItemsByURL(provider.Parameters.Metadata, subpath))
	}

	if clusterProviderItem.CertManager == nil {
		return eputils.GetError("errCertMgrCfg")
	}
	files.Files = append(files.Files, generateFileItemsByURL(clusterProviderItem.CertManager.URL, "cert-manager"))

	return nil
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_cluster_manifest := input_cluster_manifest(in)

	output_docker_images := output_docker_images(outp)
	output_files := output_files(outp)

	if input_ep_params == nil || input_ep_params.Kitconfig == nil {
		log.Errorln("Failed to find Kitconfigs for ClusterAPI cluster.")
		return eputils.GetError("errParam")
	}

	infra_provider, err := capiutils.GetInfraProvider(input_ep_params.Kitconfig)
	if err != nil {
		log.Errorln(err)
		return eputils.GetError("errCAPIProvider")
	}

	// TODO remove once refactor was done
	_ = infra_provider

	if input_cluster_manifest == nil || input_cluster_manifest.CapiClusterProviders == nil {
		log.Errorln("Failed to find manifest for ClusterAPI cluster.")
		return eputils.GetError("errCAPIManifest")
	}

	if err = getDockerImagesList(input_cluster_manifest.CapiClusterProviders, infra_provider, output_docker_images); err != nil {
		return err
	}
	if err = getProviderBinaryList(input_cluster_manifest.CapiClusterProviders, infra_provider, output_files); err != nil {
		log.Errorln(err)
		return err
	}

	kindprovider, err := cutils.GetClusterManifest(input_cluster_manifest, "kind")
	if err != nil {
		log.Errorln("Failed to find kind info for ClusterAPI cluster.")
		return eputils.GetError("errCAPIKindLost")
	}

	if err = getManagementClusterBinaryList(kindprovider, output_docker_images, output_files); err != nil {
		log.Errorln(err)
		return eputils.GetError("errMgmtCluster")
	}

	log.Debugf("%v", output_docker_images)
	log.Debugf("%v", output_files)

	return nil
}
