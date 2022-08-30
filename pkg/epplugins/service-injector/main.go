/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package serviceinjector

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	papi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	docker "github.com/intel/edge-conductor/pkg/eputils/docker"
)

func getFileFromList(filelist *papi.Files, url string) (*papi.FilesItems0, error) {
	for _, file := range filelist.Files {
		if file.URL == url {
			return file, nil
		}
	}
	return nil, eputils.GetError("errNotInList")
}

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_kitcfg := input_ep_params.Kitconfig
	input_downloadfiles := input_downloadfiles(in)
	input_serviceconfig := input_serviceconfig(in)

	output_serviceconfig := output_serviceconfig(outp)

	if input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP == "" || input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort == "" {
		return eputils.GetError("errNoServerPort")
	}

	_, err := docker.GetAuthConf(input_kitcfg.Parameters.GlobalSettings.ProviderIP,
		input_kitcfg.Parameters.GlobalSettings.RegistryPort,
		input_kitcfg.Parameters.Customconfig.Registry.User,
		input_kitcfg.Parameters.Customconfig.Registry.Password)
	if err != nil {
		log.Warnf("Docker get auth failed")
		return err
	}

	for _, service := range input_serviceconfig.Components {
		log.Infof("Injector service %s", service.Name)
		if service.Type == "repo" || service.Type == "dce" {
			log.Infof("No injection for services %s of \"repo\" type.", service.Name)
		} else {
			file, err := getFileFromList(input_downloadfiles, service.URL)
			if err != nil {
				log.Warnf("Service %s is not downloaded.", service.Name)
				continue
			} else {
				log.Infof("Service %s is available at %s", service.Name, file.Mirrorurl)
				service.URL = file.Mirrorurl
			}
		}
		if (service.Type != "repo" && len(service.Chartoverride) > 0) && (service.Type != "dce" && len(service.Chartoverride) > 0) {
			file, err := getFileFromList(input_downloadfiles, service.Chartoverride)
			if err != nil {
				log.Warnf("Service chart override %s is not downloaded.", service.Chartoverride)
				break
			} else {
				log.Infof("Service chart override is available at %s", file.Mirrorurl)
				service.Chartoverride = file.Mirrorurl
			}
		}
		for i, wanted_image := range service.Images {
			if strings.Index(wanted_image, "/") > 0 {
				registryUrl := fmt.Sprintf("%s:%s", input_kitcfg.Parameters.GlobalSettings.ProviderIP, input_kitcfg.Parameters.GlobalSettings.RegistryPort)
				newTag, err := docker.TagImageToLocal(wanted_image, registryUrl)
				if err != nil {
					return err
				}
				log.Infof("Image %s is available at %s", wanted_image, newTag)
				service.Images[i] = newTag

			}
		}

		output_serviceconfig.Components = append(
			output_serviceconfig.Components, service)
		log.Debugf("service append : %s ", service.Name)
	}

	return nil
}
