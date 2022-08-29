/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package serviceparser

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "service-parser"
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
func output_serviceconfig(outp *eputils.SchemaMapData) *pluginapi.Serviceconfig {
	return (*outp)[__name("serviceconfig")].(*pluginapi.Serviceconfig)
}

//nolint:deadcode,unused
func output_downloadfiles(outp *eputils.SchemaMapData) *pluginapi.Files {
	return (*outp)[__name("downloadfiles")].(*pluginapi.Files)
}

//nolint:deadcode,unused
func output_docker_images(outp *eputils.SchemaMapData) *pluginapi.Images {
	return (*outp)[__name("docker-images")].(*pluginapi.Images)
}

func init() {
	eputils.AddSchemaStruct(__name("ep-params"), func() eputils.SchemaStruct { return &pluginapi.EpParams{} })
	eputils.AddSchemaStruct(__name("serviceconfig"), func() eputils.SchemaStruct { return &pluginapi.Serviceconfig{} })
	eputils.AddSchemaStruct(__name("downloadfiles"), func() eputils.SchemaStruct { return &pluginapi.Files{} })
	eputils.AddSchemaStruct(__name("docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })

	Input[__name("ep-params")] = &pluginapi.EpParams{}
	Output[__name("serviceconfig")] = &pluginapi.Serviceconfig{}
	Output[__name("downloadfiles")] = &pluginapi.Files{}
	Output[__name("docker-images")] = &pluginapi.Images{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
