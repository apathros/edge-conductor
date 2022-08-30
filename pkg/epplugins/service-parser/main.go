/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package serviceparser

import (
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/yaml"

	papi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
)

func addFile(filelist *papi.Files, url, hash, hashtype, subfolder string) {
	url_origin := strings.Replace(url, path.Base(url), "", 1)

	file := papi.FilesItems0{
		URL:      url,
		Hash:     hash,
		Hashtype: hashtype,
		Urlreplacement: &papi.FilesItems0Urlreplacement{
			New:    subfolder,
			Origin: url_origin,
		},
	}
	filelist.Files = append(filelist.Files, &file)
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_kitcfg := input_ep_params.Kitconfig
	output_serviceconfig := output_serviceconfig(outp)
	output_downloadfiles := output_downloadfiles(outp)
	output_docker_images := output_docker_images(outp)

	origin_configfiles := input_kitcfg.Components.Manifests
	selectorList := input_kitcfg.Components.Selector

	all_services := papi.Serviceconfig{}
	for _, origin_configfile := range origin_configfiles {
		serviceconfig := papi.Serviceconfig{}
		err := eputils.LoadSchemaStructFromYamlFile(&serviceconfig, origin_configfile)
		if err != nil {
			log.Errorln(err)
			return err
		}
		all_services.Components = append(all_services.Components, serviceconfig.Components...)
	}
	// Select services according to select service list and service support cluster
	for _, selector := range selectorList {
		if service := findService(selector.Name, all_services.Components); service != nil {
			service, err := UpdateComponentsCustomCfg(service, selectorList)
			if err != nil {
				return err
			}
			if isSupported(input_kitcfg.Cluster.Provider, service.SupportedClusters) {
				output_serviceconfig.Components = append(
					output_serviceconfig.Components, service)
				for _, wanted_image := range service.Images {
					log.Infof("Docker image %s will be pulled for %s.", wanted_image, service.Name)
					output_docker_images.Images = append(output_docker_images.Images,
						&papi.ImagesItems0{Name: service.Name, URL: wanted_image})
				}
			} else {
				log.Warningf("Service %s is not supported on %s cluster.", service.Name, input_kitcfg.Cluster.Provider)
			}
		}
	}

	// Generate download file list
	for _, service := range output_serviceconfig.Components {
		if service.Type == "repo" || service.Type == "dce" {
			continue
		}

		subfolder := path.Join(service.Type, service.Name)
		if service.Type == "helm" && service.URL == "" {
			if service.Helmrepo != "" && service.Chartname != "" {
				ref, err := repo.FindChartInRepoURL(
					service.Helmrepo, service.Chartname,
					service.Chartversion, "", "", "",
					getter.All(&cli.EnvSettings{}),
				)

				if err != nil {
					log.Errorln(err)
					return err
				}

				service.URL = ref
			} else {
				log.Errorf("Helm repo or chart name is empty")
				return eputils.GetError("errHelmEmpty")
			}
		}
		addFile(output_downloadfiles, service.URL, service.Hash, service.Hashtype, subfolder)
		if len(service.Chartoverride) > 0 {
			addFile(output_downloadfiles, service.Chartoverride, "", "", subfolder)
		}
	}

	return nil
}

func findService(serviceName string, serviceList []*papi.Component) *papi.Component {
	for _, service := range serviceList {
		if service.Name == serviceName {
			return service
		}
	}
	return nil
}

func isSupported(clusterProvider string, supportClusterList []string) bool {
	for _, supportCluster := range supportClusterList {
		if clusterProvider == supportCluster || supportCluster == "default" {
			return true
		}
	}
	return false
}

func UpdateComponentsCustomCfg(service *papi.Component, selectServiceList []*papi.KitconfigComponentsSelectorItems0) (*papi.Component, error) {
	for _, selectService := range selectServiceList {
		if selectService.Name == service.Name {
			if selectService.OverrideYaml != "" {
				serviceMap, err := eputils.ConvertStructToMap(service)
				if err != nil {
					log.Errorf("Convert struct failed: %v", err)
					return service, eputils.GetError("errConvert")
				}
				overrideMap := map[string]interface{}{}
				err = yaml.Unmarshal([]byte(selectService.OverrideYaml), &overrideMap)
				if err != nil {
					log.Errorf("override unmarshal error: %v", err)
					return service, eputils.GetError("errOverride")
				}
				mergedMap := eputils.MergeMaps(serviceMap, overrideMap)
				err = eputils.ConvertSchemaStruct(mergedMap, service)
				if err != nil {
					log.Errorf("convert to service error: %v", err)
					return service, eputils.GetError("errService")
				}
			}
		}
	}
	return service, nil
}
