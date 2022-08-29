/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package kindparser

import (
	pluginapi "ep/pkg/api/plugins"
	eputils "ep/pkg/eputils"
	epplugin "ep/pkg/plugin"
)

var (
	Name   = "kind-parser"
	Input  = eputils.NewSchemaMapData()
	Output = eputils.NewSchemaMapData()
)

//nolint:unparam,deadcode,unused
func __name(n string) string {
	return Name + "." + n
}

//nolint:deadcode,unused
func input_cluster_manifest(in eputils.SchemaMapData) *pluginapi.Clustermanifest {
	return in[__name("cluster-manifest")].(*pluginapi.Clustermanifest)
}

//nolint:deadcode,unused
func output_docker_images(outp *eputils.SchemaMapData) *pluginapi.Images {
	return (*outp)[__name("docker-images")].(*pluginapi.Images)
}

//nolint:deadcode,unused
func output_files(outp *eputils.SchemaMapData) *pluginapi.Files {
	return (*outp)[__name("files")].(*pluginapi.Files)
}

func init() {
	eputils.AddSchemaStruct(__name("cluster-manifest"), func() eputils.SchemaStruct { return &pluginapi.Clustermanifest{} })
	eputils.AddSchemaStruct(__name("docker-images"), func() eputils.SchemaStruct { return &pluginapi.Images{} })
	eputils.AddSchemaStruct(__name("files"), func() eputils.SchemaStruct { return &pluginapi.Files{} })

	Input[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
	Output[__name("docker-images")] = &pluginapi.Images{}
	Output[__name("files")] = &pluginapi.Files{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
