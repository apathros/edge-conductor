/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package rkeinjector

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "rke-injector"
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

//nolint:deadcode,unused
func input_files(in eputils.SchemaMapData) *pluginapi.Files {
	return in[__name("files")].(*pluginapi.Files)
}

//nolint:deadcode,unused
func output_rkeconfig(outp *eputils.SchemaMapData) *pluginapi.Filecontent {
	return (*outp)[__name("rkeconfig")].(*pluginapi.Filecontent)
}

//nolint:deadcode,unused
func output_docker_images(outp *eputils.SchemaMapData) *pluginapi.Images {
	return (*outp)[__name("docker-images")].(*pluginapi.Images)
}

func init() {
	eputils.AddSchemaStruct(__name("ep-params"), func() eputils.SchemaStruct { return &pluginapi.EpParams{} })
	eputils.AddSchemaStruct(__name("docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })
	eputils.AddSchemaStruct(__name("files"), func() eputils.SchemaStruct { return &pluginapi.Files{} })
	eputils.AddSchemaStruct(__name("rkeconfig"), func() eputils.SchemaStruct { return &pluginapi.Filecontent{} })
	eputils.AddSchemaStruct(__name("docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })

	Input[__name("ep-params")] = &pluginapi.EpParams{}
	Input[__name("docker-images")] = &pluginapi.Images{}
	Input[__name("files")] = &pluginapi.Files{}
	Output[__name("rkeconfig")] = &pluginapi.Filecontent{}
	Output[__name("docker-images")] = &pluginapi.Images{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
