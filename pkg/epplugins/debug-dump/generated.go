/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package debugdump

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "debug-dump"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

//nolint:deadcode,unused
func input_nodes(in eputils.SchemaMapData) *pluginapi.Nodes {
	return in[__name("nodes")].(*pluginapi.Nodes)
}

//nolint:deadcode,unused
func input_docker_images(in eputils.SchemaMapData) *pluginapi.Images {
	return in[__name("docker-images")].(*pluginapi.Images)
}

//nolint:deadcode,unused
func input_local_docker_images(in eputils.SchemaMapData) *pluginapi.Images {
	return in[__name("local-docker-images")].(*pluginapi.Images)
}

func init() {
	eputils.AddSchemaStruct(__name("nodes"), func() eputils.SchemaStruct { return &pluginapi.Nodes{} })
	eputils.AddSchemaStruct(__name("docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })
	eputils.AddSchemaStruct(__name("local-docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })

	Input[__name("nodes")] = &pluginapi.Nodes{}
	Input[__name("docker-images")] = &pluginapi.Images{}
	Input[__name("local-docker-images")] = &pluginapi.Images{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
