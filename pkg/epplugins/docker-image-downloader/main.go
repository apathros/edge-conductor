/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package dockerimagedownloader

import (
	eputils "ep/pkg/eputils"
	docker "ep/pkg/eputils/docker"
	restfulcli "ep/pkg/eputils/restfulcli"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	DayZeroCertFilePath = "cert/pki/ca.pem"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {
	input_ep_params := input_ep_params(in)
	input_docker_images := input_docker_images(in)

	if input_docker_images.Images == nil || len(input_docker_images.Images) == 0 {
		return nil
	}

	if input_ep_params.Kitconfig == nil || input_ep_params.Kitconfig.Parameters == nil || input_ep_params.Kitconfig.Parameters.GlobalSettings == nil || input_ep_params.Kitconfig.Parameters.Customconfig == nil {
		return eputils.GetError("errKitCfgParameter")
	}

	imagesFromHost, err := docker.GetHostImages()
	if err != nil {
		return err
	}

	auth, err := docker.GetAuthConf(input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP,
		input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.User,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.Password)
	if err != nil {
		return err
	}

	var newImages []string
	var images_download []string
	var images_push_to_harbor []string
	var url string
	if eputils.CheckCmdline(input_ep_params.Cmdline, "force-download") {
		for _, img := range input_docker_images.Images {
			url = img.URL
			images_download = append(images_download, url)
			images_push_to_harbor = append(images_push_to_harbor, url)
		}
	} else {
		for _, img := range input_docker_images.Images {
			url = img.URL
			images_push_to_harbor = append(images_push_to_harbor, url)

			tmpStr := strings.TrimPrefix(url, "docker.io/")
			if _, ok := (*imagesFromHost)[tmpStr]; ok {
				continue
			}
			images_download = append(images_download, url)
		}
	}
	for _, v := range images_download {
		log.Infof("Pull image %s", v)
		if err := docker.ImagePull(v, nil); err != nil {
			return err
		}
	}

	if newImages, err = restfulcli.MapImageURLCreateHarborProject(input_ep_params.Kitconfig.Parameters.GlobalSettings.ProviderIP,
		input_ep_params.Kitconfig.Parameters.GlobalSettings.RegistryPort,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.User,
		input_ep_params.Kitconfig.Parameters.Customconfig.Registry.Password, images_push_to_harbor); err != nil {
		return err
	}

	for _, img := range newImages {
		prefixUrl := img

		newTag, err := docker.TagImageToLocal(prefixUrl, auth.ServerAddress)
		if err != nil {
			return err
		}
		log.Infof("Push %s to %s", prefixUrl, newTag)
		if err := docker.ImagePush(newTag, auth); err != nil {
			return err
		}

	}

	return nil
}
