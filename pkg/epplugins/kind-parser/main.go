/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */
// Template auto-generated once, maintained by plugin owner.

package kindparser

import (
	papi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	cutils "ep/pkg/eputils/conductorutils"

	log "github.com/sirupsen/logrus"
)

func PluginMain(in eputils.SchemaMapData, outp *eputils.SchemaMapData) error {

	input_cluster_manifest := input_cluster_manifest(in)

	output_docker_images := output_docker_images(outp)
	output_files := output_files(outp)

	provider, err := cutils.GetClusterManifest(input_cluster_manifest, "kind")
	if err != nil {
		log.Errorln("Failed to find manifest for KIND cluster.")
		return err
	}

	kindImage, err := cutils.GetImageFromProvider(provider, "img_node")
	if err != nil {
		log.Errorln("Failed to find container image for KIND node.")
		return err
	}
	kindHAProxy, err := cutils.GetImageFromProvider(provider, "img_haproxy")
	if err != nil {
		log.Errorln("Failed to find container image for KIND haproxy.")
		return err
	}
	kindBin, kindSHA256, err := cutils.GetBinaryFromProvider(provider, "kindtool")
	if err != nil {
		log.Errorln("Failed to find binary for KIND.")
		return err
	}

	output_docker_images.Images = []*papi.ImagesItems0{
		{
			Name: "kind",
			URL:  kindImage,
		},
		{
			Name: "kindhaproxy",
			URL:  kindHAProxy,
		},
	}

	output_files.Files = []*papi.FilesItems0{
		{
			URL:      kindBin,
			Hash:     kindSHA256,
			Hashtype: "sha256",
			Urlreplacement: &papi.FilesItems0Urlreplacement{
				New:    "binary",
				Origin: eputils.GetBaseUrl(kindBin),
			},
		},
	}

	return nil
}
