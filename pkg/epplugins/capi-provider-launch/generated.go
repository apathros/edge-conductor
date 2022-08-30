/*
* Copyright (c) 2022 Intel Corporation.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

// Auto generated, do not modify.

package capiproviderlaunch

import (
	pluginapi "github.com/intel/edge-conductor/pkg/api/plugins"
	eputils "github.com/intel/edge-conductor/pkg/eputils"
	epplugin "github.com/intel/edge-conductor/pkg/plugin"
)

var (
	Name   = "capi-provider-launch"
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
func input_cluster_manifest(in eputils.SchemaMapData) *pluginapi.Clustermanifest {
	return in[__name("cluster-manifest")].(*pluginapi.Clustermanifest)
}

//nolint:deadcode,unused
func input_files(in eputils.SchemaMapData) *pluginapi.Files {
	return in[__name("files")].(*pluginapi.Files)
}

func init() {
	eputils.AddSchemaStruct(__name("ep-params"), func() eputils.SchemaStruct { return &pluginapi.EpParams{} })
	eputils.AddSchemaStruct(__name("cluster-manifest"), func() eputils.SchemaStruct { return &pluginapi.Clustermanifest{} })
	eputils.AddSchemaStruct(__name("files"), func() eputils.SchemaStruct { return &pluginapi.Files{} })

	Input[__name("ep-params")] = &pluginapi.EpParams{}
	Input[__name("cluster-manifest")] = &pluginapi.Clustermanifest{}
	Input[__name("files")] = &pluginapi.Files{}

	epplugin.RegisterPlugin(Name, &Input, &Output, PluginMain)
}
