/*
* Copyright (c) 2021 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package conductorutils

import (
	papi "ep/pkg/api/plugins"
	"ep/pkg/eputils"

	log "github.com/sirupsen/logrus"
)

func GetClusterManifest(manifest *papi.Clustermanifest, name string) (*papi.ClustermanifestClusterProvidersItems0, error) {
	providers := manifest.ClusterProviders
	for _, p := range providers {
		if p.Name == name {
			return p, nil
		}
	}
	log.Warningf("Manifest for %s not found.", name)
	return nil, eputils.GetError("errManifest")
}

func GetImageFromProvider(provider *papi.ClustermanifestClusterProvidersItems0, name string) (string, error) {
	images := provider.Images
	for _, i := range images {
		if i.Name == name {
			return i.RepoTag, nil
		}
	}
	log.Warningf("Image %s not found.", name)
	return "", eputils.GetError("errImage")
}

func GetImageListFromProvider(provider *papi.ClustermanifestClusterProvidersItems0) []string {
	images := provider.Images
	var imagelist []string
	for _, i := range images {
		imagelist = append(imagelist, i.RepoTag)
	}
	return imagelist
}

func GetBinaryFromProvider(provider *papi.ClustermanifestClusterProvidersItems0, name string) (string, string, error) {
	binaries := provider.Binaries
	for _, b := range binaries {
		if b.Name == name {
			return b.URL, b.Sha256, nil
		}
	}
	log.Warningf("Binary %s not found.", name)
	return "", "", eputils.GetError("errBinary")
}

func GetResourceValueFromProvider(provider *papi.ClustermanifestClusterProvidersItems0, name string) (string, error) {
	resources := provider.Resources
	for _, r := range resources {
		if r.Name == name {
			return r.Value, nil
		}
	}
	log.Warningf("Resource %s not found.", name)
	return "", eputils.GetError("errResource")
}
