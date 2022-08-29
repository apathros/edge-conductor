/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package dockerimagedownloader

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "docker-image-downloader"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

//nolint:deadcode,unused
func input_ep_params(in eputils.SchemaMapData) *pluginapi.EpParams {
	return in[__name("ep-params")].(*pluginapi.EpParams)
}

//nolint:deadcode,unused
func input_docker_images(in eputils.SchemaMapData) *pluginapi.Images {
	return in[__name("docker-images")].(*pluginapi.Images)
}

func init() {
	eputils.AddSchemaStruct(__name("ep-params"), func() eputils.SchemaStruct { return &pluginapi.EpParams{} })
	eputils.AddSchemaStruct(__name("docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })

	Input[__name("ep-params")] = &pluginapi.EpParams{}
	Input[__name("docker-images")] = &pluginapi.Images{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
